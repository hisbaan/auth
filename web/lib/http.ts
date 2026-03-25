export function withQuery(path: string, params: Record<string, string | undefined>) {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value) {
      search.set(key, value);
    }
  }

  const query = search.toString();
  return query ? `${path}?${query}` : path;
}

export function withEncodedRequest(path: string, rawQuery: string) {
	return withQuery(path, { request: rawQuery });
}

export function parseEncodedRequest(value: string | null | undefined) {
	if (!value) {
		return null;
	}

	try {
		return new URLSearchParams(value);
	} catch {
		return null;
	}
}
