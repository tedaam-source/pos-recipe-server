"use client";

import { useEffect, useState } from "react";

interface AppEvent {
    id: string;
    message_id: string;
    status: string;
    error?: string;
    created_at: string;
}

export default function EventsPage() {
    const [events, setEvents] = useState<AppEvent[]>([]);
    const [loading, setLoading] = useState(true);

    // Simple Action Trigger
    const triggerAction = async (action: string) => {
        if (!confirm(`Trigger ${action}?`)) return;
        await fetch(`/admin/actions/${action}`, { method: 'POST' });
        alert("Action triggered");
    };

    useEffect(() => {
        fetch('/admin/events?limit=50')
            .then(res => res.json())
            .then(data => {
                setEvents(data || []);
                setLoading(false);
            });
    }, []);

    if (loading) return <div className="p-8">Loading...</div>;

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold">Recent Events</h1>
                <div className="space-x-2">
                    <button onClick={() => triggerAction('renew-watch')} className="bg-gray-200 text-gray-800 px-3 py-1 rounded text-sm hover:bg-gray-300">Renew Watch</button>
                    <button onClick={() => triggerAction('resync')} className="bg-gray-200 text-gray-800 px-3 py-1 rounded text-sm hover:bg-gray-300">Resync 24h</button>
                </div>
            </div>

            <div className="bg-white shadow rounded-lg overflow-hidden">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Message ID</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Error</th>
                        </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                        {events.map(event => (
                            <tr key={event.id}>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    {new Date(event.created_at).toLocaleString()}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">{event.message_id}</td>
                                <td className="px-6 py-4 whitespace-nowrap">
                                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${event.status === 'ok' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                                        {event.status}
                                    </span>
                                </td>
                                <td className="px-6 py-4 text-sm text-red-500">{event.error}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
