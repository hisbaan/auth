import Link from "next/link";

import { SiteHeader } from "@/components/layout/site-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { authorizeConsentAction } from "@/lib/actions";
import { withAuth } from "@/lib/auth";
import { API_BASE_URL, DOCS_URL } from "@/lib/config";
import { parseEncodedRequest, withEncodedRequest } from "@/lib/http";
import { getOIDCScopeDetail } from "@/lib/oidc";
import { getAuthorizeClientInfo } from "@/lib/sdk";

type AuthorizePageProps = {
	searchParams: Promise<{ request?: string }>;
};

export default async function AuthorizePage({ searchParams }: AuthorizePageProps) {
	const params = await searchParams;
	const request = params.request ?? "";
	const authorizeParams = parseEncodedRequest(request);

	if (!request || !authorizeParams) {
		return (
			<div className="min-h-screen">
				<main className="mx-auto flex w-full max-w-4xl justify-center px-4 py-10 sm:px-6">
					<Card className="w-full max-w-xl">
						<CardHeader>
							<CardTitle>Invalid authorize request</CardTitle>
							<CardDescription>
								We could not read the authorization request.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<p className="text-sm text-muted-foreground">
								Developer? <Link href={DOCS_URL} className="text-foreground underline underline-offset-4">Get started</Link>
							</p>
						</CardContent>
					</Card>
				</main>
			</div>
		);
	}

	const loginRedirect = withEncodedRequest("/authorize", request);
	const { cookieHeader } = await withAuth({
		loginRedirect: `/login?next=${encodeURIComponent(loginRedirect)}`,
	});

	const clientId = authorizeParams.get("client_id") ?? "";
	const scope = authorizeParams.get("scope") ?? "";
	const redirectURI = authorizeParams.get("redirect_uri") ?? "";
	const client = clientId ? await getAuthorizeClientInfo(cookieHeader, clientId) : null;
	const requestedScopes = scope.split(/\s+/).map((value) => value.trim()).filter(Boolean);
	const denyHref = `${API_BASE_URL}/authorize/deny?request=${encodeURIComponent(request)}`;

	if (!client || !redirectURI) {
		return (
			<div className="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(182,215,168,0.32),_transparent_28%),linear-gradient(180deg,_#fbf9f4,_#f3efe4)]">
				<SiteHeader />
				<main className="mx-auto flex w-full max-w-5xl justify-center px-4 py-10 sm:px-6">
					<Card className="w-full max-w-xl border-stone-300/70 bg-white/85 shadow-lg backdrop-blur">
						<CardHeader>
							<CardTitle>Unable to continue</CardTitle>
							<CardDescription>
								The client or redirect details are missing from this request.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<Button variant="outline" asChild>
								<Link href="/account">Back to account</Link>
							</Button>
						</CardContent>
					</Card>
				</main>
			</div>
		);
	}

	return (
		<div className="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(182,215,168,0.32),_transparent_28%),linear-gradient(180deg,_#fbf9f4,_#f3efe4)] text-stone-900">
			<SiteHeader />
			<main className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-10 sm:px-6 lg:grid-cols-[1.1fr_0.9fr]">
				<section className="space-y-5">
					<div className="space-y-3">
						<Badge variant="secondary" className="bg-stone-900 text-stone-50">Consent required</Badge>
						<h1 className="font-[family-name:var(--font-space-grotesk)] text-4xl font-semibold tracking-tight">{client.name} wants to connect to your account</h1>
						<p className="max-w-xl text-sm leading-6 text-stone-600">
							Review the information this app is requesting before you continue.
						</p>
					</div>

					<Card className="border-stone-300/70 bg-white/80 shadow-lg backdrop-blur">
						<CardHeader>
							<CardTitle>Requested access</CardTitle>
							<CardDescription>
								These permissions will be saved for future sign-ins until you revoke access.
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-3">
							{requestedScopes.map((requestedScope) => {
								const details = getOIDCScopeDetail(requestedScope);

								return (
									<div key={requestedScope} className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3">
										<div className="font-medium text-stone-900">{details.title}</div>
										<p className="mt-1 text-sm text-stone-600">{details.description}</p>
									</div>
								);
							})}
						</CardContent>
					</Card>
				</section>

				<section>
					<Card className="border-stone-300/70 bg-white/88 shadow-xl backdrop-blur">
						<CardHeader>
							<CardTitle>Connection details</CardTitle>
							<CardDescription>
								Confirm where you will be sent after approving access.
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-5">
							<div className="space-y-1 text-sm">
								<div className="text-stone-500">Application</div>
								<div className="font-medium">{client.name}</div>
							</div>
							<div className="space-y-1 text-sm">
								<div className="text-stone-500">Client ID</div>
								<div className="font-mono text-xs break-all text-stone-700">{client.id}</div>
							</div>
							<div className="space-y-1 text-sm">
								<div className="text-stone-500">Redirect URI</div>
								<div className="font-mono text-xs break-all text-stone-700">{redirectURI}</div>
							</div>

							<form action={authorizeConsentAction} className="space-y-3">
								<input type="hidden" name="request" value={request} />
								<input type="hidden" name="client_id" value={client.id} />
								<input type="hidden" name="scope" value={scope} />
								<Button type="submit" className="w-full bg-stone-900 text-stone-50 hover:bg-stone-800">Allow access</Button>
							</form>

							<Button type="button" variant="outline" asChild className="w-full border-stone-300 bg-transparent">
								<Link href={denyHref}>Deny</Link>
							</Button>
						</CardContent>
					</Card>
				</section>
			</main>
		</div>
	);
}
