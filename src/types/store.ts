import { DownloadIcon } from "lucide-react";
import ScissorsIcon from "@/components/store/icons/Scissors";

// ---------- Config Store ---------- //

export const store = {
  name: "SVG Piggies",
  currency: {
    code: "USD",
    label: "US",
  },
};


// ---------- Style Store ---------- //

export const design = {
  radius: {
    default: "0.5rem",
  },
  spacing: {
    section: "4rem",
    card: "1rem",
  },
  gap: {
    default: "1rem",
    grid: "1.5rem",
  },
} as const;

// ---------------------------------------- //


export interface TopProductTypesMenuSettings {
  productTypesCount: 4 | 5 | 6;
  showAllProducts: boolean;
}

export const topProductTypesMenuSettings: TopProductTypesMenuSettings = {
  productTypesCount: 6,
  showAllProducts: true,
};

export const currency = {
  code: "USD",
  label: "US",
}

export const fileFormats = {
  svg: "SVG",
  png: "PNG",
  dxf: "DXF",
  pdf: "PDF",
} as const satisfies Record<string, string>;

export type FileFormatsType = keyof typeof fileFormats;

export const productFeatures = {
  DigitalDownload: {
    label: "Digital Download",
    Icon: DownloadIcon,
  },
  Supplies: {
    label: "Supplies for making crafts",
    Icon: ScissorsIcon,
  },
} as const;

export type ProductFeatureType = keyof typeof productFeatures;
