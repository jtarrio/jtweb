function getLanguage() {
    let elem: HTMLElement | null = document.body;
    while (elem != null && elem.lang == '') {
        elem = elem.parentElement;
    }
    let lang = elem ? elem.lang : '';
    let underline = lang.indexOf('_');
    if (underline > 0) return lang.substring(0, underline);
    return lang;
}

export function getTemplate(name: string) {
    let templates = Templates[getLanguage()];
    if (!templates) templates = Templates['en'];
    return templates[name];
}

export function formatDate(date: string) {
    let d = new Date(date);
    switch (getLanguage()) {
        case 'es':
            return (d.getDate() +
                ' de ' + ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
                    'julio', 'agosto', 'setiembre', 'octubre', 'noviembre', 'diciembre'][d.getMonth() - 1] +
                ' de ' + d.getFullYear() +
                ' a las ' + String(d.getHours()).padStart(2, '0') +
                ':' + String(d.getMinutes()).padStart(2, '0'));
        case 'gl':
            return (d.getDate() +
                ' de ' + ['xaneiro', 'febreiro', 'marzo', 'abril', 'maio', 'xuño',
                    'xullo', 'agosto', 'setembro', 'outubro', 'novembro', 'decembro'][d.getMonth() - 1] +
                ' de ' + d.getFullYear() +
                ' ás ' + String(d.getHours()).padStart(2, '0') +
                ':' + String(d.getMinutes()).padStart(2, '0'));
        case 'en':
        default:
            return (['January', 'February', 'March', 'April', 'May', 'June',
                'July', 'August', 'September', 'October', 'November', 'December'][d.getMonth() - 1] +
                ' ' + d.getDate() +
                ', ' + d.getFullYear() +
                ' at ' + String(d.getHours()).padStart(2, '0') +
                ':' + String(d.getMinutes()).padStart(2, '0'));
    }
}

const Templates = {
    'en': {
        'comments': `
            <h1 jtvar="singular_count">1 comment</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comments</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
    `,
        'comment': `
            <p>By <jtvar author></jtvar> on <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
    `,
        'commentform': `
            (Comment form)
    `,
    },
    'gl': {
        'comments': `
            <h1 jtvar="singular_count">1 comentario</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comentarios</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
    `,
        'comment': `
            <p>Por <jtvar author></jtvar> o <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
    `,
        'commentform': `
            (Formulario)
    `,
    },
    'es': {
        'comments': `
            <h1 jtvar="singular_count">1 comentario</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comentarios</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
    `,
        'comment': `
            <p>Por <jtvar author></jtvar> el <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
    `,
        'commentform': `
            (Formulario)
    `,
    },
}
