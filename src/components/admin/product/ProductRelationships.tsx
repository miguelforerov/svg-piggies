"use client"

import { useEffect, useRef, useState, type ReactNode } from "react";
import { ArrowDown, ArrowUp, GripVertical, Plus, RefreshCw, Trash2 } from "lucide-react";
import { DragDropProvider } from "@dnd-kit/react";
import { isSortable, useSortable } from "@dnd-kit/react/sortable";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Product } from "@/components/admin/product/ProductFormContainer";
import { getProductImages } from "@/lib/product-images";
import {
  getAdminProducts,
  getProductRelationships,
  orderedRelatedProductIds,
  replaceProductRelationships,
  searchAvailableProducts,
} from "@/lib/product-relationships";

function RelatedProductImage({ productId, title }: { productId: string; title: string }) {
  const [url, setUrl] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getProductImages(productId)
      .then((images) => {
        if (!cancelled) setUrl((images.find((image) => image.isPrimary) ?? images[0])?.imageUrl ?? null);
      })
      .catch(() => { /* Images are optional in this list. */ });
    return () => { cancelled = true; };
  }, [productId]);

  return url
    ? <img src={url} alt="" className="size-14 shrink-0 rounded-md bg-muted object-cover" />
    : <div className="flex size-14 shrink-0 items-center justify-center rounded-md bg-muted text-xs text-muted-foreground" aria-label={`No image for ${title}`}>No image</div>;
}

function SortableRelatedProduct({ id, index, disabled, children }: {
  id: string; index: number; disabled: boolean; children: ReactNode;
}) {
  const { ref, handleRef, isDragSource, isDropTarget } = useSortable({ id, index, disabled });
  return <li ref={ref} className={`flex min-w-0 flex-wrap items-center gap-3 rounded-lg border p-3 sm:flex-nowrap ${isDragSource ? "opacity-60 ring-2 ring-primary" : ""} ${isDropTarget ? "border-primary ring-2 ring-primary" : ""}`}>
    <Button ref={handleRef} type="button" variant="ghost" size="icon-sm" disabled={disabled} className="touch-none cursor-grab active:cursor-grabbing" aria-label={`Drag to reorder related product ${index + 1}`} title="Drag to reorder"><GripVertical /></Button>
    {children}
  </li>;
}

export function ProductRelationships({ productId }: { productId: string }) {
  const [products, setProducts] = useState<Product[]>([]);
  const [savedIds, setSavedIds] = useState<string[]>([]);
  const [draftIds, setDraftIds] = useState<string[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [loaded, setLoaded] = useState(false);
  const [saving, setSaving] = useState(false);
  const savingRef = useRef(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const dirty = draftIds.length !== savedIds.length || draftIds.some((id, index) => id !== savedIds[index]);

  useEffect(() => {
    let cancelled = false;
    Promise.all([getAdminProducts(), getProductRelationships(productId)])
      .then(([loadedProducts, relationships]) => {
        if (cancelled) return;
        const ids = orderedRelatedProductIds(relationships);
        setProducts(loadedProducts);
        setSavedIds(ids);
        setDraftIds(ids);
        setLoaded(true);
        setError(null);
      })
      .catch((cause) => {
        if (!cancelled) setError(cause instanceof Error ? cause.message : "Could not load related products.");
      })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [productId]);

  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const productsById = new Map(products.map((product) => [product.id, product]));
  const results = searchAvailableProducts(products, query, productId, draftIds);

  async function refresh() {
    if (dirty || savingRef.current) return;
    setLoading(true);
    setError(null);
    setNotice(null);
    try {
      const [loadedProducts, relationships] = await Promise.all([
        getAdminProducts(), getProductRelationships(productId),
      ]);
      const ids = orderedRelatedProductIds(relationships);
      setProducts(loadedProducts);
      setSavedIds(ids);
      setDraftIds(ids);
      setLoaded(true);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not refresh related products.");
    } finally {
      setLoading(false);
    }
  }

  async function save() {
    if (!dirty || savingRef.current || loading || !loaded) return;
    savingRef.current = true;
    setSaving(true);
    setError(null);
    setNotice(null);
    const requestedIds = [...draftIds];
    try {
      await replaceProductRelationships(productId, requestedIds);
      const confirmed = orderedRelatedProductIds(await getProductRelationships(productId));
      if (confirmed.length !== requestedIds.length || confirmed.some((id, index) => id !== requestedIds[index])) {
        throw new Error("The saved relationships differ from your selection. Refresh and review them.");
      }
      setSavedIds(confirmed);
      setDraftIds(confirmed);
      setNotice("Related products saved.");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not save related products.");
    } finally {
      savingRef.current = false;
      setSaving(false);
    }
  }

  function move(index: number, offset: number) {
    const target = index + offset;
    if (savingRef.current || target < 0 || target >= draftIds.length) return;
    setDraftIds((current) => {
      const next = [...current];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
    setNotice(null);
  }

  return <section className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm" aria-label="Related Products">
    <div className="flex flex-wrap items-center justify-between gap-3">
      <div><h3 className="text-lg font-semibold">Related Products</h3><p className="text-sm text-muted-foreground">Choose the products and their order in the storefront.</p></div>
      <Button type="button" variant="outline" size="sm" disabled={dirty || saving || loading} onClick={() => void refresh()}><RefreshCw /> Refresh</Button>
    </div>

    {loading && <p className="mt-5 text-sm text-muted-foreground" role="status">Loading related products...</p>}
    {error && <p className="mt-4 rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive" role="alert">{error}</p>}
    {notice && <p className="mt-4 rounded-lg border bg-muted p-3 text-sm" role="status">{notice}</p>}

    {!loading && loaded && <>
      <div className="mt-5">
        <label htmlFor={`related-search-${productId}`} className="text-sm font-medium">Add related products</label>
        <Input id={`related-search-${productId}`} className="mt-2" type="search" value={query} disabled={saving} placeholder="Search by title or slug" onChange={(event) => setQuery(event.target.value)} />
        {query.trim() && <div className="mt-2 max-h-64 space-y-1 overflow-y-auto rounded-lg border p-2" aria-label="Available products">
          {results.length === 0 && <p className="p-2 text-sm text-muted-foreground">No available products match your search.</p>}
          {results.map((product) => <div key={product.id} className="flex min-w-0 items-center justify-between gap-2 rounded-md p-2 hover:bg-muted">
            <div className="min-w-0"><p className="truncate text-sm font-medium">{product.title}</p><p className="truncate text-xs text-muted-foreground">{product.slug} · {product.id}</p></div>
            <Button type="button" variant="secondary" size="sm" disabled={saving} aria-label={`Add ${product.title}`} onClick={() => {
              setDraftIds((current) => current.includes(product.id) || product.id === productId ? current : [...current, product.id]);
              setNotice(null);
            }}><Plus /> Add</Button>
          </div>)}
        </div>}
      </div>

      <div className="mt-6">
        <h4 className="font-medium">Current selection</h4>
        {draftIds.length === 0 && <p className="mt-3 rounded-lg bg-muted p-6 text-center text-sm text-muted-foreground">No related products yet. Search above to add some.</p>}
        {draftIds.length > 0 && <DragDropProvider onDragEnd={(event) => {
          if (event.canceled || savingRef.current) return;
          const source = event.operation.source;
          if (!isSortable(source) || source.initialIndex === source.index) return;
          setDraftIds((current) => {
            const next = [...current];
            const [moved] = next.splice(source.initialIndex, 1);
            next.splice(source.index, 0, moved);
            return next;
          });
          setNotice(null);
        }}><ol className="mt-3 space-y-2">
          {draftIds.map((id, index) => {
            const product = productsById.get(id);
            return <SortableRelatedProduct key={id} id={id} index={index} disabled={saving}>
              {product && <RelatedProductImage productId={id} title={product.title} />}
              <div className="min-w-0 flex-1"><p className="truncate text-sm font-medium">{product?.title ?? "Product unavailable"}</p><p className="break-all text-xs text-muted-foreground">{product?.slug ?? id}</p><p className="text-xs text-muted-foreground">Position {index + 1}</p></div>
              <div className="flex flex-wrap gap-1">
                <Button type="button" variant="ghost" size="icon-sm" disabled={saving || index === 0} aria-label={`Move ${product?.title ?? id} up`} onClick={() => move(index, -1)}><ArrowUp /></Button>
                <Button type="button" variant="ghost" size="icon-sm" disabled={saving || index === draftIds.length - 1} aria-label={`Move ${product?.title ?? id} down`} onClick={() => move(index, 1)}><ArrowDown /></Button>
                <Button type="button" variant="outline" size="sm" disabled={saving} aria-label={`Remove ${product?.title ?? id}`} onClick={() => { setDraftIds((current) => current.filter((candidate) => candidate !== id)); setNotice(null); }}><Trash2 /> Remove</Button>
              </div>
            </SortableRelatedProduct>;
          })}
        </ol></DragDropProvider>}
      </div>

      {dirty && <div className="mt-5 flex flex-wrap items-center justify-between gap-3 rounded-lg bg-muted p-3"><p className="text-sm">Unsaved related product changes. Save or discard before leaving this page.</p><div className="flex gap-2"><Button type="button" variant="outline" disabled={saving} onClick={() => { setDraftIds([...savedIds]); setNotice(null); setError(null); }}>Discard</Button><Button type="button" disabled={saving} onClick={() => void save()}>{saving ? "Saving..." : "Save related products"}</Button></div></div>}
    </>}
  </section>;
}
