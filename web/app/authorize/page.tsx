import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
			<div className="min-h-screen">
				<main className="mx-auto flex w-full max-w-5xl justify-center px-4 py-10 sm:px-6">
					<Card className="w-full max-w-xl">
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
		<div className="min-h-screen">
			<main className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-10 sm:px-6 lg:grid-cols-[1.1fr_0.9fr]">
				<section className="space-y-5">
					<div className="space-y-3">
						<Badge variant="secondary">Consent required</Badge>
						<h1 className="font-[family-name:var(--font-space-grotesk)] text-4xl font-semibold tracking-tight">{client.name} wants to connect to your account</h1>
						<p className="max-w-xl text-sm leading-6 text-muted-foreground">
							Review the information this app is requesting before you continue.
						</p>
					</div>

					<Card>
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
									<div key={requestedScope} className="rounded-xl border border-border bg-muted/40 px-4 py-3">
										<div className="font-medium text-foreground">{details.title}</div>
										<p className="mt-1 text-sm text-muted-foreground">{details.description}</p>
									</div>
								);
							})}
						</CardContent>
					</Card>
				</section>

				<section>
					<Card>
						<CardHeader>
							<CardTitle>Connection details</CardTitle>
							<CardDescription>
								Confirm where you will be sent after approving access.
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-5">
							<div className="space-y-1 text-sm">
								<div className="text-muted-foreground">Application</div>
								<div className="font-medium">{client.name}</div>
							</div>
							<div className="space-y-1 text-sm">
								<div className="text-muted-foreground">Client ID</div>
								<div className="font-mono text-xs break-all text-muted-foreground">{client.id}</div>
							</div>
							<div className="space-y-1 text-sm">
								<div className="text-muted-foreground">Redirect URI</div>
								<div className="font-mono text-xs break-all text-muted-foreground">{redirectURI}</div>
							</div>

							<form action="/authorize/complete" method="post" className="space-y-3">
								<input type="hidden" name="request" value={request} />
								<input type="hidden" name="client_id" value={client.id} />
								<input type="hidden" name="scope" value={scope} />
								<Button type="submit" className="w-full">Allow access</Button>
							</form>

							<Button type="button" variant="outline" asChild className="w-full">
								<Link href={denyHref}>Deny</Link>
							</Button>
						</CardContent>
					</Card>
				</section>
			</main>
		</div>
	);
}
