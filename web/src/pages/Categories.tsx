import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { categoryApi } from '../lib/category-api';
import { Category } from '../lib/types';
import Layout from '../components/Layout';

export default function Categories() {
  const { activeHousehold } = useHousehold();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [name, setName] = useState('');
  const [icon, setIcon] = useState('');

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [editIcon, setEditIcon] = useState('');

  const fetchCategories = async () => {
    if (!activeHousehold) return;
    try {
      const data = await categoryApi.list(activeHousehold.id);
      setCategories(data);
      setError('');
    } catch {
      setError('Erro ao carregar categorias');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setLoading(true);
    fetchCategories();
  }, [activeHousehold?.id]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeHousehold || !name.trim()) return;
    try {
      await categoryApi.create(activeHousehold.id, { name: name.trim(), icon: icon.trim() });
      setName('');
      setIcon('');
      setError('');
      await fetchCategories();
    } catch {
      setError('Erro ao criar categoria. Verifique se o nome já existe.');
    }
  };

  const startEdit = (cat: Category) => {
    setEditingId(cat.id);
    setEditName(cat.name);
    setEditIcon(cat.icon);
    setError('');
  };

  const handleUpdate = async () => {
    if (!activeHousehold || !editingId || !editName.trim()) return;
    try {
      await categoryApi.update(activeHousehold.id, editingId, {
        name: editName.trim(),
        icon: editIcon.trim(),
      });
      setEditingId(null);
      setError('');
      await fetchCategories();
    } catch {
      setError('Erro ao atualizar categoria.');
    }
  };

  const handleDelete = async (id: string, catName: string) => {
    if (!activeHousehold) return;
    if (!confirm(`Excluir a categoria "${catName}"?`)) return;
    try {
      await categoryApi.delete(activeHousehold.id, id);
      setError('');
      await fetchCategories();
    } catch {
      setError('Erro ao excluir categoria. Pode estar em uso.');
    }
  };

  if (!activeHousehold) {
    return (
      <Layout>
        <p className="text-center text-gray-500">
          Selecione ou crie uma residência primeiro.
        </p>
      </Layout>
    );
  }

  return (
    <Layout>
      <h2 className="mb-6 text-xl font-bold text-gray-900">Categorias</h2>

      {error && (
        <div className="mb-4 rounded bg-red-50 p-3 text-sm text-red-600">
          {error}
        </div>
      )}

      {/* Create form */}
      <form
        onSubmit={handleCreate}
        className="mb-6 flex items-end gap-3 rounded-lg bg-white p-4 shadow"
      >
        <div className="flex-1">
          <label className="mb-1 block text-sm font-medium text-gray-700">
            Nome
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Ex: Alimentação"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            maxLength={100}
            required
          />
        </div>
        <div className="w-28">
          <label className="mb-1 block text-sm font-medium text-gray-700">
            Ícone
          </label>
          <input
            type="text"
            value={icon}
            onChange={(e) => setIcon(e.target.value)}
            placeholder="🍔"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            maxLength={50}
          />
        </div>
        <button
          type="submit"
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Adicionar
        </button>
      </form>

      {/* List */}
      {loading ? (
        <p className="text-gray-500">Carregando...</p>
      ) : categories.length === 0 ? (
        <p className="text-center text-gray-400">Nenhuma categoria cadastrada.</p>
      ) : (
        <div className="overflow-hidden rounded-lg bg-white shadow">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Ícone
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Nome
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                  Ações
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {categories.map((cat) => (
                <tr key={cat.id}>
                  <td className="whitespace-nowrap px-6 py-4 text-lg">
                    {editingId === cat.id ? (
                      <input
                        type="text"
                        value={editIcon}
                        onChange={(e) => setEditIcon(e.target.value)}
                        className="w-16 rounded border border-gray-300 px-2 py-1 text-sm"
                        maxLength={50}
                      />
                    ) : (
                      cat.icon || '—'
                    )}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                    {editingId === cat.id ? (
                      <input
                        type="text"
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        className="w-48 rounded border border-gray-300 px-2 py-1 text-sm"
                        maxLength={100}
                        autoFocus
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleUpdate();
                          if (e.key === 'Escape') setEditingId(null);
                        }}
                      />
                    ) : (
                      cat.name
                    )}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-right text-sm">
                    {editingId === cat.id ? (
                      <span className="flex justify-end gap-2">
                        <button
                          onClick={handleUpdate}
                          className="text-green-600 hover:text-green-800"
                        >
                          ✓ Salvar
                        </button>
                        <button
                          onClick={() => setEditingId(null)}
                          className="text-gray-400 hover:text-gray-600"
                        >
                          Cancelar
                        </button>
                      </span>
                    ) : (
                      <span className="flex justify-end gap-3">
                        <button
                          onClick={() => startEdit(cat)}
                          className="text-blue-600 hover:text-blue-800"
                        >
                          Editar
                        </button>
                        <button
                          onClick={() => handleDelete(cat.id, cat.name)}
                          className="text-red-600 hover:text-red-800"
                        >
                          Excluir
                        </button>
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Layout>
  );
}
