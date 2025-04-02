import pb from '$lib/pocketbase/client'
import type { Album } from '$lib/pocketbase/types';
import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import type { ClientResponseError } from 'pocketbase';

export const load: PageServerLoad = async ({ url }) => {
    console.log(`Route: /`)

    try {
        let res = await pb.collection<Album>('album').getList(1, 50, {
            expand: 'artist_id'
        });

        return {
            albums: res.items // I'm assuming you meant res.items instead of results
        };
    } catch (error) {
        const pbError = error as ClientResponseError;
        console.log(pbError)
        // Check for 403 Forbidden (authentication) error
        if (pbError.status === 403) {
            // Redirect to login page
            throw redirect(303, '/login');
        }

        // Handle other errors
        console.error('Error fetching albums:', pbError);
        return {
            post: [],
            error: pbError.message || 'Failed to load albums'
        };
    }
};