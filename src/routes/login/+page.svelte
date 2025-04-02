<script lang="ts">
	import { A, Button, Heading, Input, Label } from 'flowbite-svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
	$inspect(data);
</script>

<div class="flex min-h-full flex-col justify-center px-6 py-12 lg:px-8">
	<div class="sm:mx-auto sm:w-full sm:max-w-sm">
		<img class="mx-auto h-10 w-auto" src="/favicon.png" alt="Your Company" />
		<Heading tag="h2" class="mt-10 text-center"
			>Sign in</Heading
		>
	</div>

	<div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
		{#if data.methods && data.methods.password.enabled}
			<form class="space-y-6" action="?/login" method="POST">
				<div>
					<Label for="email" class="mb-2">Email address</Label>
						<Input
							type="email"
							name="email"
							id="email"
							autocomplete="email"
							required
							class=""
						/>
				</div>

				<div>
					<div class="flex items-center justify-between mb-2">
						<Label for="password" class="">Password</Label>
						<div class="text-sm">
							<A href="#" class=""
								>Forgot password?</A
							>
						</div>
					</div>
						<Input
							type="password"
							name="password"
							id="password"
							autocomplete="current-password"
							required
							class=""
						/>
				</div>

				<div>
					<Button
						type="submit"
						class="w-full"
						>Sign in</Button
					>
				</div>
			</form>
		{/if}
	</div>
</div>
