export default {
  async fetch(request, env) {
    const url = new URL(request.url);

    if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/health')) {
      const target = new URL(url.pathname + url.search, env.API_TARGET);
      const headers = new Headers(request.headers);
      headers.set('Host', new URL(env.API_TARGET).host);

      try {
        return await fetch(target.toString(), {
          method: request.method,
          headers,
          body: request.body,
        });
      } catch {
        return new Response(JSON.stringify({ error: 'backend unavailable' }), {
          status: 502,
          headers: { 'Content-Type': 'application/json' },
        });
      }
    }

    return new Response('Not found', { status: 404 });
  },
};
