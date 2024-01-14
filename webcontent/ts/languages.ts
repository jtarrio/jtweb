import * as Data from './languagedata';

export { MessageType } from './languagedata';

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

export function getTemplate() {
    let template = Data.Templates[getLanguage()];
    if (!template) template = Data.Templates['en'];
    return template;
}


export function getMessage(msg: Data.MessageType) {
    let msgs = Data.Messages[getLanguage()];
    if (!msgs) msgs = Data.Messages['en'];
    let out = msgs[msg];
    return out === undefined ? '[unknown message: ' + Data.MessageType[msg] + ']' : out;
}

export function formatDate(date: string) {
    let d = new Date(date);
    switch (getLanguage()) {
        case 'es':
            return (d.getDate() +
                ' de ' + ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
                    'julio', 'agosto', 'setiembre', 'octubre', 'noviembre', 'diciembre'][d.getMonth()] +
                ' de ' + d.getFullYear() +
                ' a las ' + String(d.getHours()).padStart(2, '0') +
                ':' + String(d.getMinutes()).padStart(2, '0'));
        case 'gl':
            return (d.getDate() +
                ' de ' + ['xaneiro', 'febreiro', 'marzo', 'abril', 'maio', 'xuño',
                    'xullo', 'agosto', 'setembro', 'outubro', 'novembro', 'decembro'][d.getMonth()] +
                ' de ' + d.getFullYear() +
                ' ás ' + String(d.getHours()).padStart(2, '0') +
                ':' + String(d.getMinutes()).padStart(2, '0'));
        case 'en':
        default:
            return (['January', 'February', 'March', 'April', 'May', 'June',
                'July', 'August', 'September', 'October', 'November', 'December'][d.getMonth()] +
                ' ' + d.getDate() +
                ', ' + d.getFullYear() +
                ' at ' + String(d.getHours()).padStart(2, '0') +
                ':' + String(d.getMinutes()).padStart(2, '0'));
    }
}

