import { formatDate } from "./languages";

export type TmplMap = object;

// <p>Some text <jv>varName</jv></p>
// <p>Parse date <jv date>varName</jv></p>
// <p>Insert as html <jv html>varName</jv></p>
// <p>Nested vars <jv>varName.field1.field2</jv>
// <a jv-href="varName">...content...</a>
// <jv-if cond="varName">...content...</jv-if>
// <jv-if not cond="varName">...content...</jv-if>
// <jv-for items="varName" item="itemVarName">...content...</jv-for>
// <jv-for items="varName" item="itemVarName" index="indexVarName">...content...</jv-for>

export default function applyTemplate(element: Element, map: TmplMap) {
    if (element.nodeName == 'JV') {
        replaceElement(element, map);
    } else if (element.nodeName == 'JV-IF') {
        doIf(element, map);
    } else if (element.nodeName == 'JV-FOR') {
        doFor(element, map);
    } else {
        replaceAttributes(element, map);
    }
}

function getMapValue(name: any, map: TmplMap): any {
    let scope: any = map;
    if (name === null || name === undefined) return undefined;
    while (true) {
        if (name == '') return scope;
        if ('object' !== typeof scope) return undefined;
        let dot = name.indexOf('.')
        let index = dot == -1 ? name : name.substring(0, dot);
        name = dot == -1 ? '' : name.substring(dot + 1);
        let newScope = undefined;
        if (Array.isArray(scope)) {
            let indexNum = Number(index);
            if (!Number.isNaN(indexNum)) {
                newScope = scope[indexNum];
            }
        }
        if (newScope === undefined) newScope = scope[index];
        scope = newScope;
    }
}

function applyTemplateToChildren(parent: Element, map: TmplMap) {
    let child = parent.firstElementChild;
    while (child != null) {
        let next = child.nextElementSibling;
        applyTemplate(child, map);
        child = next;
    }
}

function replaceElement(element: Element, map: TmplMap) {
    let replacement = getMapValue(element.textContent?.trim(), map);
    if (element.hasAttribute('html')) {
        element.outerHTML = String(replacement);
    } else if (element.hasAttribute('date')) {
        element.replaceWith(formatDate(String(replacement)));
    } else {
        element.replaceWith(String(replacement));
    }
}

function doIf(element: Element, map: TmplMap) {
    let cond = Boolean(getMapValue(element.getAttribute('cond'), map));
    if (element.hasAttribute('not')) {
        cond = !cond;
    }
    if (cond) {
        applyTemplateToChildren(element, map);
        while (element.firstChild != null) element.before(element.firstChild);
    }
    element.remove();
}

function doFor(element: Element, map: TmplMap) {
    let items = getMapValue(element.getAttribute('items'), map);
    let itemVarName = element.getAttribute('item');
    let indexVarName = element.getAttribute('index');
    for (let index in items) {
        let item = items[index];
        let childMap = { ...map };
        if (itemVarName) childMap[itemVarName] = item;
        if (indexVarName) childMap[indexVarName] = index;
        let child = element.firstElementChild;
        while (child != null) {
            let next = child.nextElementSibling;
            let clone = child.cloneNode(true) as Element;
            element.before(clone);
            applyTemplate(clone, childMap);
            child = next;
        }
    }
    element.remove();
}

function replaceAttributes(element: Element, map: TmplMap) {
    let toDelete: string[] = [];
    for (let attr of element.attributes || []) {
        if (!attr.name.startsWith('jv-')) continue;
        let attrName = attr.name.substring(3);
        let varName = attr.value;
        let replacement = getMapValue(varName, map);
        toDelete.push(attr.name);
        if ('boolean' === typeof replacement) {
            if (replacement) {
                element.setAttribute(attrName, '');
            }
        } else {
            element.setAttribute(attrName, String(replacement));
        }
    }
    toDelete.forEach(name => element.removeAttribute(name));
    applyTemplateToChildren(element, map);
}

