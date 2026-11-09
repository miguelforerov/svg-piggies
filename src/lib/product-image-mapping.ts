export interface PublicImageRecord {
  id: string;
  imageUrl: string;
  altText?: string | null;
  widthPx?: number | null;
  heightPx?: number | null;
  displayOrder: number;
  isPrimary: boolean;
}

export function normalizePublicImages(records: PublicImageRecord[]) {
  return [...records]
    .sort((a, b) => a.displayOrder - b.displayOrder || a.id.localeCompare(b.id))
    .map((image) => ({
      id: image.id,
      url: image.imageUrl,
      altText: image.altText,
      widthPx: image.widthPx,
      heightPx: image.heightPx,
      displayOrder: image.displayOrder,
      isPrimary: image.isPrimary,
    }));
}

export function refreshedImageFor<T extends { id: string }>(records: T[], imageId: string | undefined, index: number): T | undefined {
  return imageId ? records.find((image) => image.id === imageId) : records[index];
}
