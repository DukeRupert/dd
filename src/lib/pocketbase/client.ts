import PocketBase from 'pocketbase';
import { PUBLIC_POCKETBASE_URL } from '$env/static/public'

const pb_url = PUBLIC_POCKETBASE_URL || 'http://127.0.0.1:8090'

const pb = new PocketBase(pb_url);

export default pb