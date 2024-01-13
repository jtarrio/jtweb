import { AdminApi } from "./api";

const Commands: { [key: string]: (params: string) => Promise<boolean> } = {
    'ApproveComment': approveComment
};

export async function doMailApprovals(): Promise<boolean> {
    if (window.location.hash == '') return false;
    let all = window.location.hash.substring(1);
    let i = all.indexOf('=');
    if (i == -1) return false;
    let cmd = all.substring(0, i);
    let fun = Commands[cmd];
    if (fun === undefined) return false;
    if (!await fun(all.substring(i + 1))) return false;
    window.location.hash = '';
    return true;
}

async function approveComment(params: string): Promise<boolean> {
    let parts = params.split(',');
    if (parts.length != 3) return false;
    let [cmd, postId, commentId] = parts;
    let visible: boolean;
    switch (cmd) {
        case 'approve': visible = true; break;
        case 'reject': visible = false; break;
        default: return false;
    }
    let ids = new Map([[postId, [commentId]]]);
    let success = true;
    try {
        let api = new AdminApi();
        await api.bulkSetVisible(ids, visible);
        document.body.textContent = `Success of ${cmd} of post ${postId} comment ${commentId}`;
    } catch (e) {
        success = false;
    }
    document.body.innerHTML = `<h1></h1><ul><li><b>Command:</b> </li><li><b>Post:</b> </li><li><b>Comment: </b></li>`;
    document.body.querySelector('h1')!.innerText = success ? 'Success' : 'Failure';
    let lis = document.body.querySelectorAll('li');
    lis[0].insertAdjacentText('beforeend', cmd);
    lis[1].insertAdjacentText('beforeend', postId);
    lis[2].insertAdjacentText('beforeend', commentId);
    return true;
}

