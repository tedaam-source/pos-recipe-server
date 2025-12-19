"use client";

import { useEffect, useState } from "react";
import { Plus, Trash, Edit, Check, X } from "lucide-react";

interface Filter {
    id: string;
    name: string;
    enabled: boolean;
    priority: number;
    gmail_query: string;
    updated_by?: string;
}

export default function FiltersPage() {
    const [filters, setFilters] = useState<Filter[]>([]);
    const [loading, setLoading] = useState(true);
    const [isEditing, setIsEditing] = useState<string | null>(null);
    const [newFilter, setNewFilter] = useState<Partial<Filter> | null>(null);

    const fetchFilters = () => {
        fetch('/admin/filters')
            .then(res => res.json())
            .then(data => {
                setFilters(data || []);
                setLoading(false);
            });
    };

    useEffect(() => {
        fetchFilters();
    }, []);

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure?")) return;
        await fetch(`/admin/filters/${id}`, { method: 'DELETE' });
        fetchFilters();
    };

    const handleSave = async (filter: Partial<Filter>) => {
        const method = filter.id ? 'PATCH' : 'POST';
        const url = filter.id ? `/admin/filters/${filter.id}` : '/admin/filters';

        // Ensure priority is number
        const payload = { ...filter, priority: Number(filter.priority) };

        const res = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (res.ok) {
            setIsEditing(null);
            setNewFilter(null);
            fetchFilters();
        } else {
            alert("Failed to save");
        }
    };

    if (loading) return <div className="p-8">Loading...</div>;

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold">Filters</h1>
                <button
                    onClick={() => setNewFilter({ name: '', gmail_query: '', priority: 100, enabled: true })}
                    className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 flex items-center gap-2"
                >
                    <Plus size={16} /> Add Filter
                </button>
            </div>

            <div className="bg-white shadow rounded-lg overflow-hidden">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Priority</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Query</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                            <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                        {newFilter && (
                            <FilterRow
                                filter={newFilter as Filter}
                                isEditing={true}
                                onSave={handleSave}
                                onCancel={() => setNewFilter(null)}
                            />
                        )}
                        {filters.map(filter => (
                            <FilterRow
                                key={filter.id}
                                filter={filter}
                                isEditing={isEditing === filter.id}
                                onEdit={() => setIsEditing(filter.id)}
                                onSave={handleSave}
                                onDelete={() => handleDelete(filter.id)}
                                onCancel={() => setIsEditing(null)}
                            />
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}

function FilterRow({ filter, isEditing, onEdit, onSave, onDelete, onCancel }: any) {
    const [data, setData] = useState(filter);

    if (isEditing) {
        return (
            <tr>
                <td className="px-6 py-4 whitespace-nowrap">
                    <input type="number" className="border rounded px-2 py-1 w-20" value={data.priority} onChange={e => setData({ ...data, priority: e.target.value })} />
                </td>
                <td className="px-6 py-4">
                    <input className="border rounded px-2 py-1 w-full" value={data.name} onChange={e => setData({ ...data, name: e.target.value })} />
                </td>
                <td className="px-6 py-4">
                    <input className="border rounded px-2 py-1 w-full" value={data.gmail_query} onChange={e => setData({ ...data, gmail_query: e.target.value })} />
                </td>
                <td className="px-6 py-4">
                    <input type="checkbox" checked={data.enabled} onChange={e => setData({ ...data, enabled: e.target.checked })} />
                </td>
                <td className="px-6 py-4 text-right space-x-2">
                    <button onClick={() => onSave(data)} className="text-green-600 hover:text-green-900"><Check size={20} /></button>
                    <button onClick={onCancel} className="text-gray-600 hover:text-gray-900"><X size={20} /></button>
                </td>
            </tr>
        )
    }

    return (
        <tr>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{filter.priority}</td>
            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{filter.name}</td>
            <td className="px-6 py-4 text-sm text-gray-500 max-w-xs truncate" title={filter.gmail_query}>{filter.gmail_query}</td>
            <td className="px-6 py-4 whitespace-nowrap">
                <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${filter.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}`}>
                    {filter.enabled ? 'Active' : 'Disabled'}
                </span>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                <button onClick={onEdit} className="text-indigo-600 hover:text-indigo-900"><Edit size={16} /></button>
                <button onClick={onDelete} className="text-red-600 hover:text-red-900"><Trash size={16} /></button>
            </td>
        </tr>
    );
}
