import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = { title: 'BlctekIP IP-Chain', description: 'AI 训练数据合规确权与交易平台' };
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) { return <html lang="zh-CN"><body>{children}</body></html>; }
