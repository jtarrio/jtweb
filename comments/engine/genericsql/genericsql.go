package genericsql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/engine"
)

func int64ToCommentId(id int64) comments.CommentId {
	return comments.CommentId(strconv.FormatInt(id, 36))
}

func commentIdToInt64(id comments.CommentId) (int64, error) {
	return strconv.ParseInt(string(id), 36, 64)
}

type GenericSqlEngine struct {
	db *sql.DB
}

func NewGenericSqlEngine(db *sql.DB) engine.Engine {
	return &GenericSqlEngine{db: db}
}

func (e *GenericSqlEngine) GetConfig(postId comments.PostId) (*engine.Config, error) {
	return doInReadTx(e, func(tx *sql.Tx) (*engine.Config, error) {
		return getConfig(tx, postId)
	})
}

func postRowFields() string {
	return `PostId, State, StateFromWeb`
}

func parsePostRow(rows *sql.Rows) (*engine.Config, error) {
	var rowPostId string
	var rowState int
	var rowStateFromWeb sql.NullInt16
	if err := rows.Scan(&rowPostId, &rowState, &rowStateFromWeb); err != nil {
		return nil, sqlError(err)
	}
	cfg := &engine.Config{
		PostId: comments.PostId(rowPostId),
		State:  engine.CommentState(rowState),
	}
	if rowStateFromWeb.Valid {
		cfg.State = engine.CommentState(rowStateFromWeb.Int16)
	}
	return cfg, nil
}

func getConfig(tx *sql.Tx, postId comments.PostId) (*engine.Config, error) {
	stmt, err := tx.Prepare(fmt.Sprintf(`SELECT %s FROM Posts WHERE PostId = ?`, postRowFields()))
	if err != nil {
		return nil, sqlError(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(string(postId))
	if err != nil {
		return nil, sqlError(err)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("post not found [%s]", postId)
	}
	cfg, err := parsePostRow(rows)
	if err != nil {
		return nil, sqlError(err)
	}
	return cfg, nil
}

func (e *GenericSqlEngine) SetConfig(newConfig, oldConfig *engine.Config) error {
	return doInWriteTxNoReturn(e, func(tx *sql.Tx) error {
		current, err := getConfig(tx, newConfig.PostId)
		if err != nil {
			return err
		}
		if (current == nil) != (oldConfig != nil) || current != oldConfig {
			return fmt.Errorf("old configuration is different from expected for post [%s]", newConfig.PostId)
		}
		stmt, err := tx.Prepare(`UPDATE Posts SET StateFromWeb = ? WHERE PostId = ?`)
		if err != nil {
			return sqlError(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(newConfig.State, newConfig.PostId)
		return sqlError(err)
	})
}

func (e *GenericSqlEngine) SetAllPostConfigs(cfg *engine.BulkConfig) error {
	return doInWriteTxNoReturn(e, func(tx *sql.Tx) error {
		knownPosts := map[engine.PostId]engine.CommentState{}
		{
			rows, err := tx.Query(`SELECT PostId, State FROM Posts`)
			if err != nil {
				return sqlError(err)
			}
			defer rows.Close()
			for rows.Next() {
				var rowPostId string
				var rowState int
				err := rows.Scan(&rowPostId, &rowState)
				if err != nil {
					return sqlError(err)
				}
				knownPosts[comments.PostId(rowPostId)] = engine.CommentState(rowState)
			}
		}
		deleteConfigs := map[engine.PostId]bool{}
		updateConfigs := map[engine.PostId]engine.CommentState{}
		addConfigs := map[engine.PostId]engine.CommentState{}
		for id := range knownPosts {
			deleteConfigs[id] = true
		}
		for _, newConfig := range cfg.Configs {
			if current, ok := knownPosts[newConfig.PostId]; ok {
				if current != newConfig.State {
					updateConfigs[newConfig.PostId] = newConfig.State
				}
				delete(deleteConfigs, newConfig.PostId)
			} else {
				addConfigs[newConfig.PostId] = newConfig.State
			}
		}
		{
			stmt, err := tx.Prepare(`INSERT INTO Posts (PostId, State) VALUES (?, ?)`)
			if err != nil {
				return sqlError(err)
			}
			defer stmt.Close()
			for id, state := range addConfigs {
				_, err := stmt.Exec(id, state)
				if err != nil {
					return sqlError(err)
				}
			}
		}
		{
			stmt, err := tx.Prepare(`UPDATE Posts SET State = ?, StateFromWeb = NULL WHERE PostId = ?`)
			if err != nil {
				return sqlError(err)
			}
			defer stmt.Close()
			for id, state := range updateConfigs {
				_, err := stmt.Exec(state, id)
				if err != nil {
					return sqlError(err)
				}
			}
		}
		{
			stmt, err := tx.Prepare(`DELETE FROM Posts WHERE PostId = ?`)
			if err != nil {
				return err
			}
			defer stmt.Close()
			for id := range deleteConfigs {
				_, err := stmt.Exec(id)
				if err != nil {
					return sqlError(err)
				}
			}
		}
		return nil
	})
}

func commentRowFields() string {
	return `PostId, CommentId, Visible, Author, Date, Text`
}

func parseCommentRow(rows *sql.Rows) (*engine.Comment, error) {
	var rowPostId string
	var rowCommentId int64
	var rowVisible bool
	var rowAuthor string
	var rowWhen time.Time
	var rowText string
	if err := rows.Scan(&rowPostId, &rowCommentId, &rowVisible, &rowAuthor, &rowWhen, &rowText); err != nil {
		return nil, sqlError(err)
	}
	return &engine.Comment{
		PostId:    comments.PostId(rowPostId),
		CommentId: int64ToCommentId(rowCommentId),
		Visible:   rowVisible,
		Author:    rowAuthor,
		When:      rowWhen,
		Text:      comments.Markdown(rowText),
	}, nil
}

func (e *GenericSqlEngine) List(postId comments.PostId, seeDrafts bool) ([]*engine.Comment, error) {
	return doInReadTx(e, func(tx *sql.Tx) ([]*engine.Comment, error) {
		stmt, err := tx.Prepare(fmt.Sprintf(`SELECT %s FROM Comments WHERE PostId = ? AND (Visible OR ?)`, commentRowFields()))
		if err != nil {
			return nil, sqlError(err)
		}
		defer stmt.Close()
		rows, err := stmt.Query(postId, seeDrafts)
		if err != nil {
			return nil, sqlError(err)
		}
		defer rows.Close()
		out := []*engine.Comment{}
		for rows.Next() {
			cmt, err := parseCommentRow(rows)
			if err != nil {
				return nil, sqlError(err)
			}
			out = append(out, cmt)
		}
		return out, nil
	})
}

func (e *GenericSqlEngine) Add(newComment *engine.NewComment) (*engine.Comment, error) {
	return doInWriteTx(e, func(tx *sql.Tx) (*engine.Comment, error) {
		stmt, err := tx.Prepare(`INSERT INTO Comments (PostId, Visible, Author, Date, Text) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return nil, sqlError(err)
		}
		defer stmt.Close()
		result, err := stmt.Exec(newComment.PostId, newComment.Visible, newComment.Author, newComment.When, newComment.Text)
		if err != nil {
			return nil, sqlError(err)
		}
		newId, err := result.LastInsertId()
		if err != nil {
			return nil, sqlError(err)
		}
		return &engine.Comment{
			PostId:    newComment.PostId,
			CommentId: int64ToCommentId(newId),
			Visible:   newComment.Visible,
			Author:    newComment.Author,
			When:      newComment.When,
			Text:      newComment.Text,
		}, nil
	})
}

func (e *GenericSqlEngine) FindComments(filter engine.CommentFilter, sort engine.Sort, limit int, start int) ([]*engine.Comment, error) {
	return doInReadTx(e, func(tx *sql.Tx) ([]*engine.Comment, error) {
		where, args := whereStrComments(filter)
		order := orderStrComments(sort)
		stmt, err := tx.Prepare(fmt.Sprintf(`SELECT %s FROM Comments WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, commentRowFields(), where, order, limit, start))
		if err != nil {
			return nil, sqlError(err)
		}
		defer stmt.Close()
		rows, err := stmt.Query(args...)
		if err != nil {
			return nil, sqlError(err)
		}
		defer rows.Close()
		out := []*engine.Comment{}
		for rows.Next() {
			cmt, err := parseCommentRow(rows)
			if err != nil {
				return nil, sqlError(err)
			}
			out = append(out, cmt)
		}
		return out, nil
	})
}

func whereStrComments(filter engine.CommentFilter) (string, []any) {
	args := []any{}
	where := []string{}
	if filter.Visible != nil {
		where = append(where, `Visible = ?`)
		args = append(args, filter.Visible)
	}
	if len(where) == 0 {
		return "TRUE", []any{}
	}
	return strings.Join(where, " AND "), args
}

func orderStrComments(sort engine.Sort) string {
	switch sort {
	default:
		return "Date DESC"
	}
}

func (e *GenericSqlEngine) FindPosts(filter engine.PostFilter, sort engine.Sort, limit int, start int) ([]*engine.Config, error) {
	return doInReadTx(e, func(tx *sql.Tx) ([]*engine.Config, error) {
		where, args := whereStrPosts(filter)
		order := orderStrPosts(sort)
		stmt, err := tx.Prepare(fmt.Sprintf(`SELECT %s FROM Posts WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, postRowFields(), where, order, limit, start))
		if err != nil {
			return nil, sqlError(err)
		}
		defer stmt.Close()
		rows, err := stmt.Query(args...)
		if err != nil {
			return nil, sqlError(err)
		}
		defer rows.Close()
		out := []*engine.Config{}
		for rows.Next() {
			cmt, err := parsePostRow(rows)
			if err != nil {
				return nil, sqlError(err)
			}
			out = append(out, cmt)
		}
		return out, nil
	})
}

func whereStrPosts(filter engine.PostFilter) (string, []any) {
	args := []any{}
	where := []string{}
	const stateEq = `((StateFromWeb IS NULL AND State = ?) OR StateFromWeb = ?)`
	const stateNotEq = `((StateFromWeb IS NULL AND State <> ?) OR StateFromWeb <> ?)`
	if filter.CommentsReadable == nil {
		if filter.CommentsWritable == nil {
			// Any
		} else if *filter.CommentsWritable {
			where = append(where, stateEq)
			args = append(args, engine.CommentsEnabled, engine.CommentsEnabled)
		} else {
			where = append(where, stateNotEq)
			args = append(args, engine.CommentsEnabled, engine.CommentsEnabled)
		}
	} else if *filter.CommentsReadable {
		if filter.CommentsWritable == nil {
			where = append(where, stateNotEq)
			args = append(args, engine.CommentsDisabled, engine.CommentsDisabled)
		} else if *filter.CommentsWritable {
			where = append(where, stateEq)
			args = append(args, engine.CommentsEnabled, engine.CommentsEnabled)
		} else {
			where = append(where, stateEq)
			args = append(args, engine.CommentsClosed, engine.CommentsClosed)
		}
	} else {
		where = append(where, stateEq)
		args = append(args, engine.CommentsDisabled, engine.CommentsDisabled)
	}
	if len(where) == 0 {
		return "TRUE", []any{}
	}
	return strings.Join(where, " AND "), args
}

func orderStrPosts(sort engine.Sort) string {
	switch sort {
	default:
		return "PostId DESC"
	}
}

func (e *GenericSqlEngine) BulkSetVisible(ids map[engine.PostId][]*engine.CommentId, visible bool) error {
	return doInWriteTxNoReturn(e, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`UPDATE Comments SET Visible = ? WHERE PostId = ? AND CommentId = ?`)
		if err != nil {
			return sqlError(err)
		}
		defer stmt.Close()
		for postId, commentIds := range ids {
			for _, commentId := range commentIds {
				cid, err := commentIdToInt64(*commentId)
				if err != nil {
					return err
				}
				_, err = stmt.Exec(visible, postId, cid)
				if err != nil {
					return sqlError(err)
				}
			}
		}
		return nil
	})
}

func (e *GenericSqlEngine) BulkUpdatePostConfigs(cfg *engine.BulkConfig) error {
	return doInWriteTxNoReturn(e, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`UPDATE Posts SET StateFromWeb = ? WHERE PostId = ?`)
		if err != nil {
			return sqlError(err)
		}
		defer stmt.Close()
		for _, cfg := range cfg.Configs {
			_, err := stmt.Exec(cfg.State, cfg.PostId)
			if err != nil {
				return sqlError(err)
			}
		}
		return nil
	})
}
