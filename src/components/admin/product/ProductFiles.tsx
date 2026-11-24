import { useEffect, useRef, useState } from "react";
import { FileUp, RefreshCw, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { AttachmentFile } from "@/components/ui/AttachmentFile";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  acceptedProductFileExtensions, deleteProductFile, getProductFiles, supportedProductFileLabel, uploadProductFile, validateProductFile,
  type AdminProductFile,
} from "@/lib/product-files";
import {
  SectionPanel,
  SectionPanelContent,
  SectionPanelDescription,
  SectionPanelFooter,
  SectionPanelHeader,
  SectionPanelTitle,
} from "@/components/ui/SectionPanel"

type Stage = "pending" | "authorizing" | "uploading" | "confirming" | "completed" | "failed";
type QueuedFile = { id: number; file: File; preview?: string | null; stage: Stage; error?: string };

const labels: Record<Stage, string> = {
  pending: "Pending", authorizing: "Requesting authorization", uploading: "Uploading to R2",
  confirming: "Confirming", completed: "Completed", failed: "Failed",
};

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KiB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MiB`;
}

function queueKey(file: File): string {
  return `${file.name}\u0000${file.size}\u0000${file.lastModified}`;
}

export function ProductFiles({ productId }: { productId: string }) {
  const [files, setFiles] = useState<AdminProductFile[]>([]);
  const [queue, setQueue] = useState<QueuedFile[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isBatchActive, setIsBatchActive] = useState(false);
  const batchActive = useRef(false);
  const activeUploads = useRef(new Set<number>());
  const nextId = useRef(0);
  const [fileToDelete, setFileToDelete] = useState<AdminProductFile | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const deleting = useRef(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  async function refresh() {
    setLoading(true);
    try {
      const result = await getProductFiles(productId);
      setFiles(result.sort((a, b) => a.displayOrder - b.displayOrder || a.id.localeCompare(b.id)));
      setLoadError(null);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Could not load files. Try again.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { void refresh(); }, [productId]);

  function changeItem(id: number, update: Partial<QueuedFile>) {
    setQueue((current) => current.map((item) => item.id === id ? { ...item, ...update } : item));
  }

  function addFiles(selected: FileList | File[]) {
    const incoming = Array.from(selected);
    setQueue((current) => {
      const seen = new Set(current.map((item) => queueKey(item.file)));
      const added: QueuedFile[] = [];
      for (const file of incoming) {
        const key = queueKey(file);
        if (seen.has(key)) continue;
        seen.add(key);
        const error = validateProductFile(file);
        added.push({ id: nextId.current++, file, stage: error ? "failed" : "pending", ...(error ? { error } : {}) });
      }
      return [...current, ...added];
    });
  }

  async function process(item: QueuedFile): Promise<boolean> {
    const validationError = validateProductFile(item.file);
    if (validationError || activeUploads.current.has(item.id)) return false;
    activeUploads.current.add(item.id);
    changeItem(item.id, { stage: "authorizing", error: undefined });
    try {
      const confirmed = await uploadProductFile(productId, item.file, (stage) => changeItem(item.id, { stage }));
      changeItem(item.id, { stage: "completed" });
      setFiles((current) => [confirmed, ...current.filter((file) => file.id !== confirmed.id)]);
      setNotice(`“${confirmed.displayName}” was uploaded.`);
      return true;
    } catch (error) {
      changeItem(item.id, { stage: "failed", error: error instanceof Error ? error.message : "Upload failed. Try again." });
      return false;
    } finally {
      activeUploads.current.delete(item.id);
    }
  }

  async function uploadPending() {
    if (batchActive.current || deleting.current || loading || loadError) return;
    batchActive.current = true;
    setIsBatchActive(true);
    setNotice(null);
    let succeeded = false;
    try {
      for (const item of queue.filter((file) => file.stage === "pending")) {
        if (await process(item)) succeeded = true;
      }
      if (succeeded) await refresh();
    } finally {
      batchActive.current = false;
      setIsBatchActive(false);
    }
  }

  async function retry(item: QueuedFile) {
    if (batchActive.current || deleting.current || loading || loadError || validateProductFile(item.file)) return;
    batchActive.current = true;
    setIsBatchActive(true);
    setNotice(null);
    try {
      if (await process(item)) await refresh();
    } finally {
      batchActive.current = false;
      setIsBatchActive(false);
    }
  }

  async function confirmDelete() {
    if (!fileToDelete || deleting.current || batchActive.current) return;
    const target = fileToDelete;
    deleting.current = true;
    setDeletingId(target.id);
    setDeleteError(null);
    try {
      await deleteProductFile(productId, target.id);
      setFiles((current) => current.filter((file) => file.id !== target.id));
      setFileToDelete(null);
      setNotice(`“${target.displayName}” was deleted.`);
      await refresh();
    } catch (error) {
      setDeleteError(error instanceof Error ? error.message : "Could not delete the file. Try again.");
    } finally {
      deleting.current = false;
      setDeletingId(null);
    }
  }

  return (
    <>
    <SectionPanel>
      <SectionPanelHeader>
        <SectionPanelTitle
          tailButton={
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={isBatchActive || deletingId !== null}
              onClick={() => void refresh()}
              aria-label="Refresh product files"
            >
              <RefreshCw className="size-4" />
              Refresh
            </Button>
          }>
            Files
        </SectionPanelTitle>
          <SectionPanelDescription>
            {supportedProductFileLabel} · maximum 50 MB per file
          </SectionPanelDescription>
      </SectionPanelHeader>

      <SectionPanelContent>
        <AttachmentFile
          accept={acceptedProductFileExtensions}
          buttonLabel="Select files"
          description="Drag files here or select them from your computer."
          disabled={isBatchActive || deletingId !== null}
          icon={<FileUp className="size-8 text-muted-foreground" />}
          onFiles={addFiles}
          onUpload={() => void uploadPending()}
          uploadDisabled={isBatchActive || deletingId !== null || loading || !!loadError || !queue.some((item) => item.stage === "pending")}
          queue={queue}
          preview={(item) => item.preview}
          statusLabel={(item) => labels[item.stage]}
          sizeLabel={(item) => formatSize(item.file.size)}
          retry={{
            show: (item) => item.stage === "failed" && !validateProductFile(item.file),
            disabled: () => isBatchActive || deletingId !== null || loading || !!loadError,
            onClick: (item) => void retry(item),
          }}
          remove={{
            show: (item) => !activeUploads.current.has(item.id),
            disabled: () => isBatchActive,
            onClick: (item) => setQueue((current) => current.filter((queued) => queued.id !== item.id)),
          }}
        />
      </SectionPanelContent>






    </SectionPanel>

    <section className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm" aria-label="Product files">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 className="text-lg font-semibold">Files</h3>
          <p className="text-sm text-muted-foreground">{supportedProductFileLabel} · maximum 50 MB per file</p>
        </div>
        <Button type="button" variant="outline" size="sm" disabled={isBatchActive || deletingId !== null} onClick={() => void refresh()} aria-label="Refresh product files">
          <RefreshCw className="size-4" /> Refresh
        </Button>
      </div>

      <AttachmentFile
        accept={acceptedProductFileExtensions}
        buttonLabel="Select files"
        description="Drag files here or select them from your computer."
        disabled={isBatchActive || deletingId !== null}
        icon={<FileUp className="size-8 text-muted-foreground" />}
        onFiles={addFiles}
        onUpload={() => void uploadPending()}
        uploadDisabled={isBatchActive || deletingId !== null || loading || !!loadError || !queue.some((item) => item.stage === "pending")}
        className="text-center"
        queue={queue}
        statusLabel={(item) => labels[item.stage]}
        sizeLabel={(item) => formatSize(item.file.size)}
        retry={{
          show: (item) => item.stage === "failed" && !validateProductFile(item.file),
          disabled: () => isBatchActive || deletingId !== null || loading || !!loadError,
          onClick: (item) => void retry(item),
        }}
        remove={{
          show: (item) => !activeUploads.current.has(item.id),
          disabled: () => isBatchActive,
          onClick: (item) => setQueue((current) => current.filter((queued) => queued.id !== item.id)),
        }}
      />

      <div className="mt-6">
        <h4 className="font-medium">Product files</h4>
        {notice && <p role="status" className="mt-3 rounded-lg border border-border p-3 text-sm text-foreground">{notice}</p>}
        {loading && <p className="mt-3 text-sm text-muted-foreground">Loading files...</p>}
        {loadError && <p role="alert" className="mt-3 rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">{loadError}</p>}
        {!loading && !loadError && files.length === 0 && <p className="mt-3 rounded-lg bg-muted p-6 text-center text-sm text-muted-foreground">No files yet. Upload the first file above.</p>}
        {files.length > 0 && <div className="mt-3 space-y-3">
          {files.map((file) => <div key={file.id} className="flex flex-wrap items-center gap-3 rounded-lg border p-3 sm:flex-nowrap">
            <div className="min-w-0 flex-1">
              <p className="truncate font-medium" title={file.displayName}>{file.displayName}</p>
              <p className="truncate text-sm text-muted-foreground" title={file.fileName}>{file.fileName} · {file.format.toUpperCase()} · {formatSize(file.fileSizeBytes)} · Order {file.displayOrder + 1}</p>
            </div>
            <Badge variant="secondary">{file.availabilityStatus}</Badge>
            <Button type="button" variant="outline" size="sm" disabled={deletingId !== null || isBatchActive}
              aria-label={`Delete ${file.displayName}`} onClick={() => { setDeleteError(null); setFileToDelete(file); }}>
              <Trash2 className="size-4" /> Delete
            </Button>
          </div>)}
        </div>}
      </div>

      <AlertDialog open={fileToDelete !== null} onOpenChange={(open) => { if (!open && !deleting.current) { setFileToDelete(null); setDeleteError(null); } }}>
        <AlertDialogContent className="bg-secondary">
          <AlertDialogHeader>
            <AlertDialogTitle>Delete product file?</AlertDialogTitle>
            <AlertDialogDescription>Delete <strong>“{fileToDelete?.displayName}”</strong> from this product? This action cannot be undone.</AlertDialogDescription>
          </AlertDialogHeader>
          {deleteError && <p role="alert" className="text-sm text-destructive">{deleteError}</p>}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deletingId !== null}>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" disabled={deletingId !== null} onClick={() => void confirmDelete()}>
              {deletingId !== null ? "Deleting..." : "Delete file"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </section>
    </>
  );
}
