export type HealthResponse = { code:string; message:string; data:{status:string}; request_id:string };
export async function getHealth(): Promise<HealthResponse | null> {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080/api/v1';
  try { const response = await fetch(`${baseUrl}/health`, { cache:'no-store' }); if (!response.ok) return null; return (await response.json()) as HealthResponse; } catch { return null; }
}
