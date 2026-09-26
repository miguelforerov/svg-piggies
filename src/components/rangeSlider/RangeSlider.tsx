import React, { useEffect, useRef, useState } from "react";
import { CATALOG_CHANGE_EVENT, navigateCatalog } from "@/lib/catalog-navigation";
import "./rangeSlider.css";

const RangeSlider = ({
  maxPriceData,
}: {
  maxPriceData: { amount: string; currencyCode: string };
}) => {
  const currencyCode = maxPriceData.currencyCode;
  const currencySymbol = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currencyCode,
  })
    .formatToParts(0)
    .find((part) => part.type === "currency")?.value ?? "$";

  const maxAmount = parseInt(maxPriceData?.amount);
  const [minValue, setMinValue] = useState(0);
  const [maxValue, setMaxValue] = useState(maxAmount);

  const rangeRef = useRef<HTMLDivElement>(null);
  const minThumbRef = useRef<HTMLDivElement>(null);
  const maxThumbRef = useRef<HTMLDivElement>(null);
  const rangeLineRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const syncPrice = () => {
      const params = new URLSearchParams(window.location.search);
      setMinValue(Number(params.get("minPrice") ?? 0));
      setMaxValue(Number(params.get("maxPrice") ?? maxAmount));
    };

    syncPrice();
    window.addEventListener("popstate", syncPrice);
    window.addEventListener(CATALOG_CHANGE_EVENT, syncPrice);
    return () => {
      window.removeEventListener("popstate", syncPrice);
      window.removeEventListener(CATALOG_CHANGE_EVENT, syncPrice);
    };
  }, [maxAmount]);

  useEffect(() => {
    updateRangeBar();
  }, [minValue, maxValue]);

  const updateRangeBar = () => {
    if (
      !rangeLineRef.current ||
      !minThumbRef.current ||
      !maxThumbRef.current ||
      !rangeRef.current
    )
      return;

    const minPercent = (minValue / maxAmount) * 100;
    const maxPercent = (maxValue / maxAmount) * 100;

    rangeLineRef.current.style.left = `${minPercent}%`;
    rangeLineRef.current.style.width = `${maxPercent - minPercent}%`;

    minThumbRef.current.style.left = `${minPercent}%`;
    maxThumbRef.current.style.left = `${maxPercent}%`;
  };

  const handleMouseDown = (thumb: "min" | "max") => (e: React.MouseEvent) => {
    e.preventDefault();

    const startX = e.clientX;
    const rangeRect = rangeRef.current?.getBoundingClientRect();
    if (!rangeRect) return;

    const rangeWidth = rangeRect.width;
    const initialMinVal = minValue;
    const initialMaxVal = maxValue;
    let nextMinValue = initialMinVal;
    let nextMaxValue = initialMaxVal;

    const handleMouseMove = (e: MouseEvent) => {
      const dx = e.clientX - startX;
      const dPercent = (dx / rangeWidth) * 100;

      if (thumb === "min") {
        const newPercent = (initialMinVal / maxAmount) * 100 + dPercent;
        const newValue = Math.max(
          0,
          Math.min(maxValue - 1, Math.round((newPercent * maxAmount) / 100))
        );
        nextMinValue = newValue;
        setMinValue(newValue);
      } else {
        const newPercent = (initialMaxVal / maxAmount) * 100 + dPercent;
        const newValue = Math.max(
          minValue + 1,
          Math.min(maxAmount, Math.round((newPercent * maxAmount) / 100))
        );
        nextMaxValue = newValue;
        setMaxValue(newValue);
      }
    };

    const handleMouseUp = () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
      priceChange(nextMinValue, nextMaxValue);
    };

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
  };

  function priceChange(min: number, max: number) {
    const params = new URLSearchParams(window.location.search);
    const appliedMin = Number(params.get("minPrice") ?? 0);
    const appliedMax = Number(params.get("maxPrice") ?? maxAmount);
    if (min === appliedMin && max === appliedMax) return;

    if (min === 0) params.delete("minPrice");
    else params.set("minPrice", String(min));
    if (max === maxAmount) params.delete("maxPrice");
    else params.set("maxPrice", String(max));
    params.delete("page");

    const query = params.toString();
    navigateCatalog(query ? `/products?${query}` : "/products");
  }

  return (
    <div className="range-slider-container">
      <div className="flex justify-between">
        <p>
          {currencySymbol}
          {minValue} {maxPriceData?.currencyCode || currencyCode}
        </p>
        <p>
          {currencySymbol}
          {maxValue} {maxPriceData?.currencyCode || currencyCode}
        </p>
      </div>

      <div className="range-slider" ref={rangeRef}>
        <div className="slider-track !bg-primary-light"></div>
        <div className="slider-range" ref={rangeLineRef}></div>
        <div
          className="slider-thumb thumb-min"
          ref={minThumbRef}
          onMouseDown={handleMouseDown("min")}
        ></div>
        <div
          className="slider-thumb thumb-max"
          ref={maxThumbRef}
          onMouseDown={handleMouseDown("max")}
        ></div>
      </div>
    </div>
  );
};

export default RangeSlider;
