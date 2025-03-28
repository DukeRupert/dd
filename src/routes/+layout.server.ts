import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ url }) => {
    // Pagination parameters
    const page = parseInt(url.searchParams.get('page') || '1', 10);
    const limit = parseInt(url.searchParams.get('limit') || '10', 10);

    // Search parameter
    const search = url.searchParams.get('search') || '';

    // Query parameters (add any additional ones you need)
    const sortBy = url.searchParams.get('sortBy') || 'createdAt';
    const sortOrder = url.searchParams.get('sortOrder') || 'desc';
    const filter = url.searchParams.get('filter') || '';

    // Calculate offset for pagination
    const offset = (page - 1) * limit;
    return {
        pagination: { page, limit, offset },
        search,
        query: { sortBy, sortOrder, filter }
    };
};