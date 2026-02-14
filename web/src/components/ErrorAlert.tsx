export default function ErrorAlert({
  message,
  onDismiss,
}: {
  message: string;
  onDismiss?: () => void;
}) {
  if (!message) return null;
  return (
    <div className="mb-4 flex items-center justify-between rounded bg-red-50 p-3 text-sm text-red-600">
      <span>{message}</span>
      {onDismiss && (
        <button
          onClick={onDismiss}
          className="ml-3 text-red-400 hover:text-red-600"
          aria-label="Fechar"
        >
          ✕
        </button>
      )}
    </div>
  );
}
