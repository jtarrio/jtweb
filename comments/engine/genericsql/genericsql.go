package genericsql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

type notFoundError struct {
	postId comments.PostId
}

func notFound(postId comments.PostId) error {
	return &notFoundError{postId: postId}
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("unknown post [%s]", e.postId)
}

type GenericSqlEngine struct {
	db *sql.DB
}

func NewGenericSqlEngine(db *sql.DB) engine.Engine {
	return &GenericSqlEngine{db: db}
}

func (e *GenericSqlEngine) GetConfig(postId comments.PostId) (*engine.Config, error) {
	tx, err := e.startReadTx()
	if err != nil {
		return nil, err
	}
	cfg, err := tx.getConfig(postId)
	tx.finishTx(err)
	return cfg, err
}

func (e *GenericSqlEngine) SetConfig(newConfig, oldConfig *engine.Config) error {
	tx, err := e.startWriteTx()
	if err != nil {
		return err
	}
	current, err := tx.getConfig(newConfig.PostId)
	if _, ok := err.(*notFoundError); !ok {
		return tx.finishTx(err)
	}
	if (current == nil) != (oldConfig != nil) || current != oldConfig {
		return tx.finishTx(fmt.Errorf("old configuration is different from expected for post [%s]", newConfig.PostId))
	}
	return tx.finishTx(tx.setConfig(newConfig))
}

func (e *GenericSqlEngine) BulkSetConfig(cfg *engine.BulkConfig) error {
	tx, err := e.startWriteTx()
	if err != nil {
		return err
	}
	return tx.finishTx(tx.bulkSetConfig(cfg))
}

func (e *GenericSqlEngine) List(postId comments.PostId, seeDrafts bool) ([]engine.Comment, error) {
	tx, err := e.startReadTx()
	if err != nil {
		return nil, err
	}
	cmts, err := tx.load(postId, seeDrafts)
	tx.finishTx(err)
	return cmts, err
}

func (e *GenericSqlEngine) Add(newComment *engine.NewComment) (*engine.Comment, error) {
	tx, err := e.startWriteTx()
	if err != nil {
		return nil, err
	}
	cmt, err := tx.add(newComment)
	tx.finishTx(err)
	return cmt, err
}

type etx struct {
	tx *sql.Tx
}

func (e *GenericSqlEngine) startReadTx() (*etx, error) {
	return e.startTx(sql.LevelReadCommitted)
}

func (e *GenericSqlEngine) startWriteTx() (*etx, error) {
	return e.startTx(sql.LevelSerializable)
}

func (e *GenericSqlEngine) startTx(level sql.IsolationLevel) (*etx, error) {
	tx, err := e.db.BeginTx(context.TODO(), &sql.TxOptions{Isolation: level})
	if err != nil {
		return nil, err
	}
	return &etx{tx: tx}, nil
}

func (tx *etx) finishTx(err error) error {
	if err != nil {
		tx.tx.Rollback()
		return err
	}
	return tx.tx.Commit()
}

func (tx *etx) getConfig(postId comments.PostId) (*engine.Config, error) {
	stmt, err := tx.tx.Prepare(`SELECT PostId, State, StateOverride FROM Posts WHERE PostId = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var rowPostId string
	var rowState int
	var rowStateOverride sql.NullInt16
	err = stmt.QueryRow(string(postId)).Scan(&rowPostId, &rowState, &rowStateOverride)
	if err == sql.ErrNoRows {
		return nil, notFound(postId)
	}
	cfg := &engine.Config{
		PostId: comments.PostId(rowPostId),
		State:  engine.CommentState(rowState),
	}
	if rowStateOverride.Valid {
		cfg.State = engine.CommentState(rowStateOverride.Int16)
	}
	return cfg, nil
}

func (tx *etx) setConfig(newConfig *engine.Config) error {
	stmt, err := tx.tx.Prepare(`UPDATE Posts SET StateOverride = ? WHERE PostId = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(newConfig.State, newConfig.PostId)
	return err
}

func (tx *etx) bulkSetConfig(cfg *engine.BulkConfig) error {
	rows, err := tx.tx.Query(`SELECT PostId, State FROM Posts`)
	if err != nil {
		return nil
	}
	knownPosts := map[engine.PostId]engine.CommentState{}
	{
		defer rows.Close()
		for rows.Next() {
			var rowPostId string
			var rowState int
			err := rows.Scan(&rowPostId, &rowState)
			if err != nil {
				return err
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
		stmt, err := tx.tx.Prepare(`INSERT INTO Posts (PostId, State) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for id, state := range addConfigs {
			_, err := stmt.Exec(id, state)
			if err != nil {
				return err
			}
		}
	}
	{
		stmt, err := tx.tx.Prepare(`UPDATE Posts SET State = ?, StateOverride = NULL WHERE PostId = ?`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for id, state := range updateConfigs {
			_, err := stmt.Exec(state, id)
			if err != nil {
				return err
			}
		}
	}
	{
		stmt, err := tx.tx.Prepare(`DELETE FROM Posts WHERE PostId = ?`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for id := range deleteConfigs {
			_, err := stmt.Exec(id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tx *etx) load(postId comments.PostId, seeDrafts bool) ([]engine.Comment, error) {
	stmt, err := tx.tx.Prepare(`SELECT PostId, CommentId, Visible, Author, Date, Text FROM Comments WHERE PostId = ? AND (Visible OR ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(postId, seeDrafts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []engine.Comment{}
	for rows.Next() {
		var rowPostId string
		var rowCommentId int64
		var rowVisible bool
		var rowAuthor string
		var rowWhen time.Time
		var rowText string
		if err := rows.Scan(&rowPostId, &rowCommentId, &rowVisible, &rowAuthor, &rowWhen, &rowText); err != nil {
			return nil, err
		}
		out = append(out, engine.Comment{
			PostId:    comments.PostId(rowPostId),
			CommentId: int64ToCommentId(rowCommentId),
			Visible:   rowVisible,
			Author:    rowAuthor,
			When:      rowWhen,
			Text:      comments.Markdown(rowText),
		})
	}
	return out, nil
}

func (tx *etx) add(newComment *engine.NewComment) (*engine.Comment, error) {
	stmt, err := tx.tx.Prepare(`INSERT INTO Comments (PostId, Visible, Author, Date, Text) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	result, err := stmt.Exec(newComment.PostId, newComment.Visible, newComment.Author, newComment.When, newComment.Text)
	if err != nil {
		return nil, err
	}
	newId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &engine.Comment{
		PostId:    newComment.PostId,
		CommentId: int64ToCommentId(newId),
		Visible:   newComment.Visible,
		Author:    newComment.Author,
		When:      newComment.When,
		Text:      newComment.Text,
	}, nil
}
