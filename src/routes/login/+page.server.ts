import pb from '$lib/pocketbase/client'
import type { PageServerLoad, Actions } from './$types';
import { redirect } from '@sveltejs/kit';
import type { ClientResponseError, } from 'pocketbase';

export const load: PageServerLoad = async ({ url }) => {
    try {
        let methods = await pb.collection('users').listAuthMethods();

        return {
            methods
        };
    } catch (error) {
        const pbError = error as ClientResponseError;
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

export const actions = {
    login: async ({ request, cookies }) => {
        const data = await request.formData();
        const email = data.get('email') as string;
        const password = data.get('password') as string;
        console.log(`Login attempt. email: ${email}, password: ${password}`)

        const authData = await pb.collection('users').authWithPassword(
            'lwilliams56@gmail.com',
            'serenity',
        );

        console.log(authData)
        console.log(pb.authStore.isValid)


        if (pb.authStore.isValid) {
            cookies.set('user', authData.record.id, { path: '/' })
            redirect(303, '/')
        }

        return { message: "Failed to authenticate" }
    }
} satisfies Actions;