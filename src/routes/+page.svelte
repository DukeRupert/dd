<script lang="ts">
    import type { PageProps } from './$types'
    import { Card, Heading, P } from 'flowbite-svelte'

    let { data }: PageProps = $props();
    $inspect(data)
    let albums = $derived(data.albums)
</script>

<h1>Albums</h1>
<p>Visit <a href="https://svelte.dev/docs/kit">svelte.dev/docs/kit</a> to read the documentation</p>
{#if albums}
{#each albums as a}
    <Card>
        <Heading tag="h3">{a.title}</Heading>
        <P>{a.expand.artist_id.name}</P>
    </Card>
{/each}
{/if}
