import { getHealth } from '../lib/api';

export default async function Home() {
  const health = await getHealth();
  const online = health?.data.status === 'ok';
  return (
    <main className="mx-auto flex min-h-screen max-w-6xl flex-col justify-center px-8 py-16">
      <p className="mb-4 text-sm font-semibold tracking-[0.3em] text-sky-300">BLCTEK IP-CHAIN</p>
      <h1 className="max-w-4xl text-5xl font-bold leading-tight md:text-7xl">数据留在本地，确权、交易与交付在可信链路中完成。</h1>
      <p className="mt-8 max-w-3xl text-lg leading-8 text-slate-300">第一阶段工程底座已建立。后续将实现 Data Agent、不可变数据版本、佣金账本与受控 P2P 交付。</p>
      <div className="mt-10 inline-flex w-fit items-center gap-3 rounded-full border border-slate-700 px-5 py-3">
        <span className={`h-3 w-3 rounded-full ${online ? 'bg-emerald-400' : 'bg-amber-400'}`} />
        <span className="text-sm text-slate-200">API {online ? '已连接' : '尚未连接'}</span>
      </div>
    </main>
  );
}
