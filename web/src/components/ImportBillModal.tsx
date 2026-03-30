import { useState, useRef, useCallback } from 'react';
import { importApi } from '../lib/import-api';
import { useAuth } from '../lib/auth';
import type {
  Category,
  Member,
  ParsedExpensePreview,
  ImportConfirmItem,
} from '../lib/types';
import Spinner from './Spinner';

interface ImportBillModalProps {
  isOpen: boolean;
  onClose: () => void;
  householdId: string;
  categories: Category[];
  members: Member[];
  onImportComplete: () => void;
}

interface EditableItem extends ParsedExpensePreview {
  selected: boolean;
  category_id: string;
  is_shared: boolean;
  paid_by: string;
  assigned_to: string;
}

function formatCurrency(cents: number): string {
  return (cents / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

export default function ImportBillModal({
  isOpen,
  onClose,
  householdId,
  categories,
  members,
  onImportComplete,
}: ImportBillModalProps) {
  const { user } = useAuth();
  const [step, setStep] = useState<'upload' | 'preview' | 'importing'>('upload');
  const [items, setItems] = useState<EditableItem[]>([]);
  const [provider, setProvider] = useState('');
  const [error, setError] = useState('');
  const [uploading, setUploading] = useState(false);
  const [dragging, setDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(
    async (file: File) => {
      setError('');
      setUploading(true);
      try {
        const result = await importApi.upload(householdId, file);
        setProvider(result.provider);
        setItems(
          result.items.map((item) => ({
            ...item,
            selected: true,
            category_id: item.suggested_category_id || '',
            is_shared: true,
            paid_by: user?.id || '',
            assigned_to: '',
          }))
        );
        setStep('preview');
      } catch {
        setError('Erro ao processar o arquivo. Verifique o formato e tente novamente.');
      } finally {
        setUploading(false);
      }
    },
    [householdId]
  );

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) handleFile(file);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (file) handleFile(file);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragging(true);
  };

  const handleDragLeave = () => setDragging(false);

  const updateItem = (index: number, patch: Partial<EditableItem>) => {
    setItems((prev) => prev.map((it, i) => (i === index ? { ...it, ...patch } : it)));
  };

  const selectedItems = items.filter((it) => it.selected);
  const allSelected = items.length > 0 && items.every((it) => it.selected);

  const toggleAll = () => {
    const next = !allSelected;
    setItems((prev) => prev.map((it) => ({ ...it, selected: next })));
  };

  const selectedTotal = selectedItems.reduce((sum, it) => sum + it.amount_cents, 0);

  const handleConfirm = async () => {
    setError('');
    setStep('importing');
    try {
      const confirmItems: ImportConfirmItem[] = selectedItems.map((it) => ({
        category_id: it.category_id,
        description: it.description,
        amount_cents: it.amount_cents,
        expense_date: it.date,
        is_shared: it.is_shared,
        paid_by: it.paid_by,
        assigned_to: it.is_shared ? '' : it.assigned_to,
      }));
      await importApi.confirm(householdId, confirmItems);
      onImportComplete();
      onClose();
    } catch {
      setError('Erro ao importar despesas. Tente novamente.');
      setStep('preview');
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-lg shadow-xl max-w-5xl w-full mx-4 max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-gray-900">
            📄 Importar Fatura
            {provider && (
              <span className="ml-2 text-sm font-normal text-gray-500">
                ({provider})
              </span>
            )}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-xl leading-none"
            aria-label="Fechar"
          >
            ✕
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {error && (
            <div className="mb-4 rounded-md bg-red-50 p-3 text-sm text-red-700">
              {error}
            </div>
          )}

          {/* Upload step */}
          {step === 'upload' && (
            <div>
              {uploading ? (
                <Spinner text="Processando arquivo..." />
              ) : (
                <div
                  onDrop={handleDrop}
                  onDragOver={handleDragOver}
                  onDragLeave={handleDragLeave}
                  onClick={() => fileInputRef.current?.click()}
                  className={`cursor-pointer rounded-lg border-2 border-dashed p-12 text-center transition ${
                    dragging
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-300 hover:border-gray-400'
                  }`}
                >
                  <div className="text-4xl mb-3">📂</div>
                  <p className="text-sm font-medium text-gray-700">
                    Arraste o arquivo da fatura aqui
                  </p>
                  <p className="mt-1 text-xs text-gray-500">
                    ou clique para selecionar (PDF, CSV, OFX)
                  </p>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".pdf,.csv,.ofx"
                    onChange={handleFileInput}
                    className="hidden"
                  />
                </div>
              )}
            </div>
          )}

          {/* Preview step */}
          {step === 'preview' && (
            <div>
              <div className="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <label className="flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={allSelected}
                    onChange={toggleAll}
                    className="rounded border-gray-300"
                  />
                  Selecionar todos ({items.length})
                </label>
                <div className="text-sm text-gray-600">
                  Selecionadas: <strong>{selectedItems.length}</strong> · Total:{' '}
                  <strong className="text-green-700">{formatCurrency(selectedTotal)}</strong>
                </div>
              </div>

              {/* Mobile cards */}
              <div className="space-y-3 sm:hidden">
                {items.map((item, idx) => (
                  <div
                    key={idx}
                    className={`rounded-lg border p-3 ${
                      item.selected ? 'border-blue-200 bg-blue-50/30' : 'border-gray-200 opacity-60'
                    }`}
                  >
                    <div className="mb-2 flex items-start gap-2">
                      <input
                        type="checkbox"
                        checked={item.selected}
                        onChange={(e) => updateItem(idx, { selected: e.target.checked })}
                        className="mt-1 rounded border-gray-300"
                      />
                      <div className="flex-1 space-y-2">
                        <input
                          type="text"
                          value={item.description}
                          onChange={(e) => updateItem(idx, { description: e.target.value })}
                          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                        />
                        <div className="grid grid-cols-2 gap-2">
                          <input
                            type="number"
                            step="0.01"
                            min="0.01"
                            value={(item.amount_cents / 100).toFixed(2)}
                            onChange={(e) => {
                              const cents = Math.round(parseFloat(e.target.value) * 100);
                              if (!isNaN(cents)) updateItem(idx, { amount_cents: cents });
                            }}
                            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                          />
                          <input
                            type="date"
                            value={item.date}
                            onChange={(e) => updateItem(idx, { date: e.target.value })}
                            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                          />
                        </div>
                        <select
                          value={item.category_id}
                          onChange={(e) => updateItem(idx, { category_id: e.target.value })}
                          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                        >
                          <option value="">Sem categoria</option>
                          {categories.map((cat) => (
                            <option key={cat.id} value={cat.id}>
                              {cat.icon} {cat.name}
                            </option>
                          ))}
                        </select>
                        <label className="flex items-center gap-2 text-sm text-gray-700">
                          <input
                            type="checkbox"
                            checked={item.is_shared}
                            onChange={(e) =>
                              updateItem(idx, { is_shared: e.target.checked, assigned_to: '' })
                            }
                            className="rounded border-gray-300"
                          />
                          Compartilhada
                        </label>
                        <select
                          value={item.paid_by}
                          onChange={(e) => updateItem(idx, { paid_by: e.target.value })}
                          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                        >
                          {members.map((m) => (
                            <option key={m.user_id} value={m.user_id}>
                              {m.user_name}
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {/* Desktop table */}
              <div className="hidden sm:block overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200 text-sm">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left">
                        <input
                          type="checkbox"
                          checked={allSelected}
                          onChange={toggleAll}
                          className="rounded border-gray-300"
                        />
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">
                        Descrição
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">
                        Valor
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">
                        Data
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">
                        Categoria
                      </th>
                      <th className="px-3 py-2 text-center text-xs font-medium uppercase text-gray-500">
                        Compartilhada
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">
                        Pago por
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {items.map((item, idx) => (
                      <tr
                        key={idx}
                        className={item.selected ? '' : 'opacity-50'}
                      >
                        <td className="px-3 py-2">
                          <input
                            type="checkbox"
                            checked={item.selected}
                            onChange={(e) =>
                              updateItem(idx, { selected: e.target.checked })
                            }
                            className="rounded border-gray-300"
                          />
                        </td>
                        <td className="px-3 py-2">
                          <input
                            type="text"
                            value={item.description}
                            onChange={(e) =>
                              updateItem(idx, { description: e.target.value })
                            }
                            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                          />
                        </td>
                        <td className="px-3 py-2">
                          <input
                            type="number"
                            step="0.01"
                            min="0.01"
                            value={(item.amount_cents / 100).toFixed(2)}
                            onChange={(e) => {
                              const cents = Math.round(
                                parseFloat(e.target.value) * 100
                              );
                              if (!isNaN(cents))
                                updateItem(idx, { amount_cents: cents });
                            }}
                            className="w-24 rounded-md border border-gray-300 px-3 py-2 text-sm"
                          />
                        </td>
                        <td className="px-3 py-2">
                          <input
                            type="date"
                            value={item.date}
                            onChange={(e) =>
                              updateItem(idx, { date: e.target.value })
                            }
                            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                          />
                        </td>
                        <td className="px-3 py-2">
                          <select
                            value={item.category_id}
                            onChange={(e) =>
                              updateItem(idx, { category_id: e.target.value })
                            }
                            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                          >
                            <option value="">Sem categoria</option>
                            {categories.map((cat) => (
                              <option key={cat.id} value={cat.id}>
                                {cat.icon} {cat.name}
                              </option>
                            ))}
                          </select>
                        </td>
                        <td className="px-3 py-2 text-center">
                          <input
                            type="checkbox"
                            checked={item.is_shared}
                            onChange={(e) =>
                              updateItem(idx, {
                                is_shared: e.target.checked,
                                assigned_to: '',
                              })
                            }
                            className="rounded border-gray-300"
                          />
                        </td>
                        <td className="px-3 py-2">
                          <select
                            value={item.paid_by}
                            onChange={(e) =>
                              updateItem(idx, { paid_by: e.target.value })
                            }
                            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                          >
                            {members.map((m) => (
                              <option key={m.user_id} value={m.user_id}>
                                {m.user_name}
                              </option>
                            ))}
                          </select>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Importing step */}
          {step === 'importing' && <Spinner text="Importando despesas..." />}
        </div>

        {/* Footer */}
        <div className="border-t border-gray-200 px-6 py-4 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
          <button
            onClick={onClose}
            className="rounded-md bg-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-300"
          >
            Cancelar
          </button>
          {step === 'preview' && (
            <button
              onClick={handleConfirm}
              disabled={selectedItems.length === 0}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Importar {selectedItems.length} despesa{selectedItems.length !== 1 ? 's' : ''}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
