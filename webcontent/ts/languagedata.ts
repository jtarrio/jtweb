
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
    'en': {
        'main': `
            <h1 jtvar="singular_count">1 comment</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comments</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
        `,
        'entry': `
            <p>By <jtvar author></jtvar> on <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
        `,
        'form': `
            <form id="commentform">
                <p>Your name: <input type="text" name="author"></p>
                <p>Comment: <textarea name="text" rows="10" cols="50"></textarea></p>
                <input type="submit" value="Submit"><input type="reset" value="Reset"><input type="button" value="Preview" id="jtPreviewButton">
            </form>
            <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
        `,
    },
    'gl': {
        'main': `
            <h1 jtvar="singular_count">1 comentario</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comentarios</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
        `,
        'entry': `
            <p>Por <jtvar author></jtvar> o <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
        `,
        'form': `
            <form id="commentform">
                <p>O teu nome: <input type="text" name="author"></p>
                <p>Comentario: <textarea name="text" rows="10" cols="50"></textarea></p>
                <input type="submit" value="Enviar"><input type="reset" value="Descartar"><input type="button" value="Previsualizar" id="jtPreviewButton">
            </form>
            <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
        `,
    },
    'es': {
        'main': `
            <h1 jtvar="singular_count">1 comentario</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comentarios</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
        `,
        'entry': `
            <p>Por <jtvar author></jtvar> el <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
        `,
        'form': `
        <form id="commentform">
            <p>Tu nombre: <input type="text" name="author"></p>
            <p>Comentario: <textarea name="text" rows="10" cols="50"></textarea></p>
            <input type="submit" value="Enviar"><input type="reset" value="Descartar"><input type="button" value="Previsualizar" id="jtPreviewButton">
        </form>
        <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
        `,
    },
}
