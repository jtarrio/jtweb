
export enum MessageType {
    ErrorPostingComment,
    CommentPostedAsDraft,
}

export const Messages = {
    'en': {
        [MessageType.ErrorPostingComment]:
            'There was an error while submitting the comment.',
        [MessageType.CommentPostedAsDraft]:
            'Your comment was submitted and will become visible when it is approved.',
    },
    'es': {
        [MessageType.ErrorPostingComment]:
            'Hubo un error enviando el comentario.',
        [MessageType.CommentPostedAsDraft]:
            'Se ha recibido tu comentario y será publicado cuando se apruebe.',
    },
    'gl': {
        [MessageType.ErrorPostingComment]:
            'Houbo un erro ao enviar o comentario.',
        [MessageType.CommentPostedAsDraft]:
            'Recibiuse o teu comentario e vai ser publicado cando se aprobe.',
    }
};

export const Templates = {
    'en': `<div class="comment">
    <jv-if cond="has_none_count"><h1>No comments</h1></jv-if>
    <jv-if cond="has_singular_count"><h1>1 comment</h1></jv-if>
    <jv-if cond="has_plural_count"><h1><jv>count</jv> comments</h1></jv-if>
    <jv-for items="comments" item="comment">
        <div class="commentByline">By <jv>comment.author</jv> on <a jv-href="comment.url" jv-name="comment.anchor"><jv date>comment.when</jv></a></div>
        <div class="commentText><jv html>comment.text</jv></div>
    </jv-for>
    <jv-if cond="can_add_comment">
        <form id="commentform">
            <h1>Would you like to write a comment?</h1>
            <div class="commentForm">
                <div>Your name or nickname: <input type="text" name="author"> (it will be published)</div>
                <div>Your comment:</div>
                <textarea name="text" id="jtComment"></textarea>
                <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
            </div>
            <div class="commentButtons">
                <input type="reset" value="Reset"><input type="submit" value="Submit"><input type="button" value="Preview" id="jtPreviewButton">
            </div>
        </form>
    </jv-if>
</div>`,
    'gl': `<div class="comment">
    <jv-if cond="has_none_count"><h1>Ningún comentario</h1></jv-if>
    <jv-if cond="has_singular_count"><h1>1 comentario</h1></jv-if>
    <jv-if cond="has_plural_count"><h1><jv>count</jv> comentarios</h1></jv-if>
    <jv-for items="comments" item="comment">
        <div class="commentByline">Por <jv>comment.author</jv> o <a jv-href="comment.url" jv-name="comment.anchor"><jv date>comment.when</jv></a></div>
        <div class="commentText><jv html>comment.text</jv></div>
    </jv-for>
    <jv-if cond="can_add_comment">
        <form id="commentform">
            <h1>Queres escribir un comentario?</h1>
            <div class="commentForm">
                <div>O teu nome ou sobrenome: <input type="text" name="author"> (hase publicar)</div>
                <div>O teu comentario:</div>
                <textarea name="text" id="jtComment"></textarea>
                <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
            </div>
            <div class="commentButtons">
                <input type="reset" value="Descartar"><input type="submit" value="Enviar"><input type="button" value="Previsualizar" id="jtPreviewButton">
            </div>
        </form>
    </jv-if>
</div>`,
    'es': `<div class="comment">
    <jv-if cond="has_none_count"><h1>Ningún comentario</h1></jv-if>
    <jv-if cond="has_singular_count"><h1>1 comentario</h1></jv-if>
    <jv-if cond="has_plural_count"><h1><jv>count</jv> comentarios</h1></jv-if>
    <jv-for items="comments" item="comment">
        <div class="commentByline">Por <jv>comment.author</jv> el <a jv-href="comment.url" jv-name="comment.anchor"><jv date>comment.when</jv></a></div>
        <div class="commentText><jv html>comment.text</jv></div>
    </jv-for>
    <jv-if cond="can_add_comment">
        <form id="commentform">
            <h1>¿Quieres escribir un comentario?</h1>
            <div class="commentForm">
                <div>Tu nombre o sobrenombre: <input type="text" name="author"> (se publicará)</div>
                <div>Tu comentario:</div>
                <textarea name="text" id="jtComment"></textarea>
                <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
            </div>
            <div class="commentButtons">
                <input type="reset" value="Descartar"><input type="submit" value="Enviar"><input type="button" value="Previsualizar" id="jtPreviewButton">
            </div>
        </form>
    </jv-if>
</div>`,
}
