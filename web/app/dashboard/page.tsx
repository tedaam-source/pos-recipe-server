"use client";

import { useEffect, useState } from "react";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
// Card imports removed as we use custom StatCard
// I will use raw divs for simplicity if components don't exist, but let's try to be clean.
// Actually, let's just write the UI inline or create a simple components file.

interface DailyStat {
    day: string;
    received: number;
    processed_ok: number;
    processed_error: number;
    last_event_at: string;
}

export default function Dashboard() {
    const [stats, setStats] = useState<DailyStat[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch('/admin/stats')
            .then(res => res.json())
            .then(data => {
                setStats(data || []);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setLoading(false);
            });
    }, []);

    if (loading) return <div className="p-8">Loading stats...</div>;

    return (
        <div className="space-y-6">
            <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>

            <div className="grid gap-4 md:grid-cols-3">
                <StatCard title="Total Received (30d)" value={stats.reduce((acc, curr) => acc + curr.received, 0)} />
                <StatCard title="Processed OK" value={stats.reduce((acc, curr) => acc + curr.processed_ok, 0)} className="text-green-600" />
                <StatCard title="Errors" value={stats.reduce((acc, curr) => acc + curr.processed_error, 0)} className="text-red-600" />
            </div>

            <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
                <h3 className="text-lg font-medium mb-4">Pipeline Activity (30 Days)</h3>
                <div className="h-[300px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={stats}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="day" tickFormatter={(val) => val.slice(5)} />
                            <YAxis />
                            <Tooltip />
                            <Legend />
                            <Bar dataKey="received" fill="#8884d8" name="Received" />
                            <Bar dataKey="processed_ok" fill="#82ca9d" name="Processed OK" />
                            <Bar dataKey="processed_error" fill="#ff8042" name="Errors" />
                        </BarChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
}

function StatCard({ title, value, className }: { title: string, value: number, className?: string }) {
    return (
        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
            <div className="text-sm font-medium text-gray-500">{title}</div>
            <div className={`text-2xl font-bold mt-2 ${className}`}>{value}</div>
        </div>
    );
}
