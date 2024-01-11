import { doAdminComments } from './admin_comments';
import { doAdminPosts } from './admin_posts';
import { initTabs } from './tabs';

window.addEventListener('load', _ => doAdmin());

function doAdmin() {
    initTabs('tabTitle', 'tabPage');
    doAdminComments(document.getElementById('tabPage-0')!);
    doAdminPosts(document.getElementById('tabPage-1')!);
}
