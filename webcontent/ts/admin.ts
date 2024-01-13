import { doAdminComments } from './admin_comments';
import { doMailApprovals } from './admin_mail';
import { doAdminPosts } from './admin_posts';
import { initTabs } from './tabs';

window.addEventListener('load', _ => doAdmin());

async function doAdmin() {
    if (await doMailApprovals()) return;

    initTabs('tabTitle', 'tabPage');
    doAdminComments(document.getElementById('tabPage-0')!);
    doAdminPosts(document.getElementById('tabPage-1')!);
}
