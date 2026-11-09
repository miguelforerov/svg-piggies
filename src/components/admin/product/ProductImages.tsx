import { useEffect, useRef, useState, type ChangeEvent, type DragEvent, type ReactNode } from "react";
import { GripVertical, ImagePlus, RefreshCw, Trash2 } from "lucide-react";
import { DragDropProvider } from "@dnd-kit/react";
import { isSortable, useSortable } from "@dnd-kit/react/sortable";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { deleteProductImage, getProductImages, ProductImageApiError, reorderProductImages, setPrimaryProductImage, updateProductImageAltText, uploadProductImage, validateImage, type ProductImageRecord } from "@/lib/product-images";

type Stage = "pending" | "authorizing" | "uploading" | "confirming" | "completed" | "error";
type QueuedImage = { id: number; file: File; preview: string; altText: string; stage: Stage; error?: string };
const labels: Record<Stage, string> = {
  pending: "Pending", authorizing: "Requesting authorization", uploading: "Uploading to R2",
  confirming: "Confirming", completed: "Completed", error: "Error",
};

function SortableImage({ image, index, disabled, children }: { image: ProductImageRecord; index: number; disabled: boolean; children: ReactNode }) {
  const { ref, handleRef, isDragSource, isDropTarget } = useSortable({ id: image.id, index, disabled });
  return <div ref={ref} className={`overflow-hidden rounded-lg border ${isDragSource ? "opacity-60 ring-2 ring-primary" : ""} ${isDropTarget ? "border-primary ring-2 ring-primary" : ""}`}>
    <div className="flex justify-end border-b bg-muted/40 px-2 py-1">
      <Button ref={handleRef} type="button" variant="ghost" size="sm" disabled={disabled} className="touch-none cursor-grab active:cursor-grabbing" aria-label={`Move ${image.originalFilename}`} title="Drag to reorder"><GripVertical className="size-4" /></Button>
    </div>
    {children}
  </div>;
}

export function ProductImages({ productId }: { productId: string }) {
  const [images, setImages] = useState<ProductImageRecord[]>([]);
  const [savedOrder, setSavedOrder] = useState<string[]>([]);
  const [draftOrder, setDraftOrder] = useState<string[]>([]);
  const [orderSaving, setOrderSaving] = useState(false);
  const orderSavingRef = useRef(false);
  const [queue, setQueue] = useState<QueuedImage[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const nextId = useRef(0);
  const previews = useRef<string[]>([]);
  const activeUploads = useRef(new Set<number>());
  const batchActive = useRef(false);
  const [isBatchActive, setIsBatchActive] = useState(false);
  const [imageToDelete, setImageToDelete] = useState<ProductImageRecord | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const deleting = useRef(false);
  const [primaryUpdatingId, setPrimaryUpdatingId] = useState<string | null>(null);
  const primaryUpdating = useRef(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [imageToEdit, setImageToEdit] = useState<ProductImageRecord | null>(null);
  const [draftAltText, setDraftAltText] = useState("");
  const [altSaving, setAltSaving] = useState(false);
  const altSavingRef = useRef(false);
  const [editError, setEditError] = useState<string | null>(null);
  const orderPending = draftOrder.length === savedOrder.length && draftOrder.some((id, index) => id !== savedOrder[index]);
  const imagesById = new Map(images.map((image) => [image.id, image]));
  const displayedImages = draftOrder.map((id) => imagesById.get(id)).filter((image): image is ProductImageRecord => image !== undefined);

  async function saveOrder() {
    if (!orderPending || orderSavingRef.current || deleting.current || batchActive.current) return;
    orderSavingRef.current = true;
    setOrderSaving(true);
    setNotice(null);
    try {
      await reorderProductImages(productId, draftOrder);
      setImages((current) => current.map((image) => ({ ...image, displayOrder: draftOrder.indexOf(image.id) })));
      setSavedOrder([...draftOrder]);
      setNotice({ kind: "success", message: "Image order saved." });
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof ProductImageApiError && error.status === 400
        ? "The image list changed. Cancel the draft, refresh the gallery, then try again."
        : error instanceof Error ? error.message : "Could not save image order. Try again." });
    } finally {
      orderSavingRef.current = false;
      setOrderSaving(false);
    }
  }

  function openAltEditor(image: ProductImageRecord) {
    setImageToEdit(image);
    setDraftAltText(image.altText ?? "");
    setEditError(null);
    setNotice(null);
  }

  async function saveAltText() {
    if (!imageToEdit || altSavingRef.current) return;
    const text = draftAltText.trim();
    if (Array.from(text).length > 500) {
      setEditError("Alternative text must be 500 characters or fewer.");
      return;
    }
    altSavingRef.current = true;
    setAltSaving(true);
    setEditError(null);
    try {
      await updateProductImageAltText(productId, imageToEdit.id, text);
      setImages((current) => current.map((image) => image.id === imageToEdit.id ? { ...image, altText: text } : image));
      setImageToEdit(null);
      setNotice({ kind: "success", message: `Alternative text for “${imageToEdit.originalFilename}” was saved.` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Could not save alternative text. Try again.";
      setEditError(message);
      setNotice({ kind: "error", message });
    } finally {
      altSavingRef.current = false;
      setAltSaving(false);
    }
  }

  async function selectPrimary(image: ProductImageRecord) {
    if (image.isPrimary || primaryUpdating.current || deleting.current) return;
    primaryUpdating.current = true;
    setPrimaryUpdatingId(image.id);
    setNotice(null);
    try {
      await setPrimaryProductImage(productId, image.id);
      setImages((current) => current.map((item) => ({ ...item, isPrimary: item.id === image.id })));
      setNotice({ kind: "success", message: `“${image.originalFilename}” is now the primary image.` });
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : "Could not set the primary image. Try again." });
    } finally {
      primaryUpdating.current = false;
      setPrimaryUpdatingId(null);
    }
  }

  async function confirmDelete() {
    if (!imageToDelete || deleting.current || primaryUpdating.current || orderPending || orderSavingRef.current) return;
    deleting.current = true;
    setDeletingId(imageToDelete.id);
    setNotice(null);
    const target = imageToDelete;
    try {
      await deleteProductImage(productId, target.id);
      setImages((current) => {
        const remaining = current.filter((image) => image.id !== target.id);
        if (target.isPrimary && remaining.length > 0) {
          const next = remaining.reduce((first, image) => image.displayOrder < first.displayOrder ? image : first);
          return remaining.map((image) => image.id === next.id ? { ...image, isPrimary: true } : image);
        }
        return remaining;
      });
      setSavedOrder((current) => current.filter((id) => id !== target.id));
      setDraftOrder((current) => current.filter((id) => id !== target.id));
      setImageToDelete(null);
      setNotice({ kind: "success", message: `“${target.originalFilename}” was deleted.` });
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : "Could not delete the image. Try again." });
    } finally {
      deleting.current = false;
      setDeletingId(null);
    }
  }

  async function refresh(force = false) {
    if (!force && (orderPending || orderSavingRef.current)) return;
    try {
      const result = await getProductImages(productId);
      result.sort((a, b) => a.displayOrder - b.displayOrder || a.id.localeCompare(b.id));
      setImages(result);
      setSavedOrder(result.map((image) => image.id));
      setDraftOrder(result.map((image) => image.id));
      setLoadError(null);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Could not load images.");
    } finally {
      setLoading(false);
    }
  }

  async function recoverImage(id: string, failedUrl: string) {
    if (failedUrl === "/images/product-placeholder.jpg") return;
    try {
      const refreshed = await getProductImages(productId);
      const replacement = refreshed.find((image) => image.id === id)?.imageUrl;
      setImages((current) => current.map((image) => image.id === id ? {
        ...image, imageUrl: replacement && replacement !== failedUrl ? replacement : "/images/product-placeholder.jpg",
      } : image));
    } catch {
      setImages((current) => current.map((image) => image.id === id ? { ...image, imageUrl: "/images/product-placeholder.jpg" } : image));
    }
  }

  useEffect(() => {
    void refresh(true);
    return () => { previews.current.forEach(URL.revokeObjectURL); };
  }, [productId]);

  function addFiles(files: FileList | File[]) {
    if (orderPending || orderSavingRef.current) return;
    const selected = Array.from(files).map((file) => {
      const preview = URL.createObjectURL(file);
      previews.current.push(preview);
      const error = validateImage(file);
      return { id: nextId.current++, file, preview, altText: "", stage: error ? "error" : "pending", error: error ?? undefined } as QueuedImage;
    });
    setQueue((current) => [...current, ...selected]);
    if (inputRef.current) inputRef.current.value = "";
  }

  function changeItem(id: number, patch: Partial<QueuedImage>) {
    setQueue((current) => current.map((item) => item.id === id ? { ...item, ...patch } : item));
  }

  async function process(item: QueuedImage, placement: { isPrimary: boolean; displayOrder: number }): Promise<boolean> {
    if (validateImage(item.file)) return false;
    if (activeUploads.current.has(item.id)) return false;
    activeUploads.current.add(item.id);
    changeItem(item.id, { stage: "authorizing", error: undefined });
    try {
      await uploadProductImage(productId, item.file, item.altText, (stage) => changeItem(item.id, { stage }), placement);
      changeItem(item.id, { stage: "completed" });
      await refresh();
      return true;
    } catch (error) {
      changeItem(item.id, { stage: "error", error: error instanceof Error ? error.message : "Upload failed. Try again." });
      return false;
    } finally {
      activeUploads.current.delete(item.id);
    }
  }

  async function uploadPending() {
    if (batchActive.current || loading || loadError || orderPending || orderSavingRef.current) return;
    batchActive.current = true;
    setIsBatchActive(true);
    let hasPrimary = images.some((image) => image.isPrimary);
    let nextOrder = Math.max(-1, ...images.map((image) => image.displayOrder)) + 1;
    try {
      for (const item of queue.filter((image) => image.stage === "pending")) {
        const succeeded = await process(item, { isPrimary: !hasPrimary, displayOrder: nextOrder });
        if (succeeded) { hasPrimary = true; nextOrder += 1; }
      }
    } finally {
      batchActive.current = false;
      setIsBatchActive(false);
    }
  }

  async function retry(item: QueuedImage) {
    if (batchActive.current || loading || loadError || orderPending || orderSavingRef.current) return;
    batchActive.current = true;
    setIsBatchActive(true);
    try {
      await process(item, { isPrimary: !images.some((image) => image.isPrimary), displayOrder: Math.max(-1, ...images.map((image) => image.displayOrder)) + 1 });
    } finally {
      batchActive.current = false;
      setIsBatchActive(false);
    }
  }

  function onInput(event: ChangeEvent<HTMLInputElement>) {
    if (event.target.files) addFiles(event.target.files);
  }

  function onDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    setDragging(false);
    addFiles(event.dataTransfer.files);
  }

  return (
    <section className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm" aria-label="Product images">
      <div className="flex items-center justify-between gap-3">
        <div><h3 className="text-lg font-semibold">Images</h3><p className="text-sm text-muted-foreground">JPEG, PNG, or WebP · maximum 10 MiB per file</p></div>
        <Button type="button" variant="outline" size="sm" disabled={orderPending || orderSaving} onClick={() => void refresh()} aria-label="Refresh product images"><RefreshCw className="size-4" /> Refresh</Button>
      </div>
      <div onDragOver={(event) => { event.preventDefault(); setDragging(true); }} onDragLeave={() => setDragging(false)} onDrop={onDrop}
        className={`mt-5 flex flex-col items-center gap-3 rounded-lg border-2 border-dashed p-8 text-center ${dragging ? "border-primary bg-muted" : "border-border"}`}>
        <ImagePlus className="size-8 text-muted-foreground" />
        <p className="text-sm">Drag images here or select files from your computer.</p>
        <input ref={inputRef} type="file" accept="image/jpeg,image/png,image/webp" multiple disabled={orderPending || orderSaving} className="sr-only" onChange={onInput} aria-label="Select product images" />
        <Button type="button" variant="secondary" disabled={orderPending || orderSaving} onClick={() => inputRef.current?.click()}>Select images</Button>
      </div>

      {queue.length > 0 && <div className="mt-5 space-y-3">
        <div className="flex items-center justify-between"><h4 className="font-medium">Selected images</h4>
          <Button type="button" disabled={isBatchActive || loading || !!loadError || orderPending || orderSaving || !queue.some((item) => item.stage === "pending")} onClick={() => void uploadPending()}>Upload pending</Button>
        </div>
        {queue.map((item) => <div key={item.id} className="flex flex-wrap items-center gap-3 rounded-lg border p-3 sm:flex-nowrap">
          <img src={item.preview} alt="Selected preview" className="size-16 rounded object-cover" />
          <div className="min-w-0 flex-1"><p className="truncate text-sm font-medium">{item.file.name}</p>
            <p className="text-xs text-muted-foreground" role="status">{labels[item.stage]}</p>
            {item.error && <p className="text-xs text-destructive" role="alert">{item.error}</p>}
            {item.stage === "pending" && <Input className="mt-2" value={item.altText} placeholder="Alternative text (optional)" aria-label={`Alternative text for ${item.file.name}`} onChange={(event) => changeItem(item.id, { altText: event.target.value })} />}
          </div>
          {item.stage === "error" && !validateImage(item.file) && <Button type="button" variant="outline" size="sm" disabled={isBatchActive || loading || !!loadError || orderPending || orderSaving} onClick={() => void retry(item)}>Retry</Button>}
        </div>)}
      </div>}

      <div className="mt-6">
        <h4 className="font-medium">Gallery</h4>
        {orderPending && <div className="mt-3 flex flex-wrap items-center gap-2"><p className="mr-auto text-sm text-muted-foreground">Unsaved image order. Save or cancel before uploading or deleting images.</p><Button type="button" variant="outline" disabled={orderSaving} onClick={() => setDraftOrder([...savedOrder])}>Cancel</Button><Button type="button" disabled={orderSaving} onClick={() => void saveOrder()}>{orderSaving ? "Saving..." : "Save order"}</Button></div>}
        {notice && <p role={notice.kind === "error" ? "alert" : "status"} className={`mt-3 rounded-lg border p-3 text-sm ${notice.kind === "error" ? "border-destructive/50 text-destructive" : "border-border text-foreground"}`}>{notice.message}</p>}
        {loading && <p className="mt-3 text-sm text-muted-foreground">Loading images...</p>}
        {loadError && <p className="mt-3 text-sm text-destructive" role="alert">{loadError}</p>}
        {!loading && !loadError && images.length === 0 && <p className="mt-3 rounded-lg bg-muted p-6 text-center text-sm text-muted-foreground">No images yet. Upload the first image above.</p>}
        {images.length > 0 && <DragDropProvider onDragEnd={(event) => {
          if (event.canceled || orderSavingRef.current || isBatchActive) return;
          const source = event.operation.source;
          if (!isSortable(source) || source.initialIndex === source.index) return;
          setDraftOrder((current) => {
            const next = [...current];
            const [moved] = next.splice(source.initialIndex, 1);
            next.splice(source.index, 0, moved);
            return next;
          });
        }}><div className="mt-3 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {displayedImages.map((image, index) => <SortableImage key={image.id} image={image} index={index} disabled={orderSaving || isBatchActive}>
            <img src={image.imageUrl} alt={image.altText ?? image.originalFilename} className="aspect-square w-full bg-muted object-contain" onError={() => void recoverImage(image.id, image.imageUrl)} />
            <div className="space-y-1 p-3 text-sm"><p className="truncate font-medium" title={image.originalFilename}>{image.originalFilename}</p>
              <p className="text-muted-foreground">{image.widthPx && image.heightPx ? `${image.widthPx} × ${image.heightPx} · ` : ""}{(image.fileSizeBytes / 1024).toFixed(1)} KiB</p>
              <p>{image.altText === "" ? "Empty alternative text" : image.altText ?? "No alternative text"}</p>
              <p className="text-muted-foreground">Order {index + 1}</p>
              <div className="mt-2 flex flex-wrap items-center gap-2">
                <Button type="button" variant="outline" size="sm" aria-label={`Edit alt text for ${image.originalFilename}`} onClick={() => openAltEditor(image)}>Edit alt text</Button>
                {image.isPrimary
                  ? <Badge variant="secondary">Primary</Badge>
                  : <Button type="button" variant="secondary" size="sm" disabled={primaryUpdatingId !== null || deletingId !== null} aria-label={`Set ${image.originalFilename} as primary`} onClick={() => void selectPrimary(image)}>{primaryUpdatingId === image.id ? "Updating..." : "Set as primary"}</Button>}
                <Button type="button" variant="outline" size="sm" disabled={deletingId !== null || primaryUpdatingId !== null || orderPending || orderSaving} aria-label={`Delete ${image.originalFilename}`} onClick={() => { setNotice(null); setImageToDelete(image); }}><Trash2 className="size-4" /> Delete</Button>
              </div>
            </div>
          </SortableImage>)}
        </div></DragDropProvider>}
      </div>
      <Dialog open={imageToEdit !== null} onOpenChange={(open) => { if (!open && !altSavingRef.current) setImageToEdit(null); }}>
        <DialogContent showDialogFooter={false} showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>Edit alt text</DialogTitle>
          </DialogHeader>
          <DialogDescription className="min-w-0 [overflow-wrap:anywhere]">Image: {imageToEdit?.originalFilename}</DialogDescription>
          <Textarea value={draftAltText} onChange={(event) => setDraftAltText(event.target.value)} aria-label="Alternative text" disabled={altSaving} />
          <p className="text-xs text-muted-foreground">Leave empty to use an empty alt attribute.</p>
          {editError && <p role="alert" className="text-sm text-destructive">{editError}</p>}
          <div className="flex flex-wrap justify-end gap-2">
            <Button type="button" variant="outline" disabled={altSaving} onClick={() => setImageToEdit(null)}>Cancel</Button>
            <Button type="button" disabled={altSaving} onClick={() => void saveAltText()}>{altSaving ? "Saving..." : "Save"}</Button>
          </div>
        </DialogContent>
      </Dialog>
      <AlertDialog
        open={imageToDelete !== null}
        onOpenChange={(open) => { if (!open && !deleting.current) setImageToDelete(null); }}
      >
        <AlertDialogContent className="bg-secondary">
          <AlertDialogHeader>
            <AlertDialogTitle>Delete product image?</AlertDialogTitle>
            <AlertDialogDescription>
              Delete <strong>“{imageToDelete?.originalFilename}”</strong> from this product?
            </AlertDialogDescription>
            <AlertDialogDescription>
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deletingId !== null}>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" disabled={deletingId !== null} onClick={() => void confirmDelete()}>
              {deletingId !== null ? "Deleting..." : "Delete image"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </section>
  );
}
