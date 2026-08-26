import React, { useState } from "react";
import { FaFacebookF, FaPinterestP, FaWhatsapp } from "react-icons/fa";
import { FiCheck, FiCopy } from "react-icons/fi";

interface SocialShareProps {
  title: string;
  url?: string;
  imageUrl?: string;
}

const SocialShare = ({ title, url, imageUrl }: SocialShareProps) => {
  const [copied, setCopied] = useState(false);

  const shareUrl =
    url || (typeof window !== "undefined" ? window.location.href : "");

  const encodedUrl = encodeURIComponent(shareUrl);
  const encodedTitle = encodeURIComponent(title);
  const encodedImageUrl = imageUrl
    ? encodeURIComponent(
        new URL(
          imageUrl,
          typeof window !== "undefined"
            ? window.location.origin
            : "http://localhost:4321",
        ).href,
      )
    : "";

  // Pinterest
  const pinterestUrl = `https://www.pinterest.com/pin/create/button/?url=${encodedUrl}&media=${encodedImageUrl}&description=${encodedTitle}`;

  // Facebook
  const facebookUrl = `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`;

  // WhatsApp
  const whatsappUrl = `https://wa.me/?text=${encodeURIComponent(
    `${title} ${shareUrl}`,
  )}`;

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);

      setTimeout(() => {
        setCopied(false);
      }, 2000);
    } catch (error) {
      console.error("Unable to copy product link:", error);
    }
  };

  return (
    <div className="flex items-center gap-2">
      {/* Pinterest */}
      <a
        href={pinterestUrl}
        target="_blank"
        rel="noopener noreferrer"
        aria-label={`Share ${title} on Pinterest`}
        className="flex h-9 w-9 items-center justify-center rounded-full border border-border transition-colors duration-300 text-text-gray hover:bg-primary hover:text-white"
      >
        <FaPinterestP size={16} />
      </a>

      {/* Facebook */}
      <a
        href={facebookUrl}
        target="_blank"
        rel="noopener noreferrer"
        aria-label={`Share ${title} on Facebook`}
        className="flex h-9 w-9 items-center justify-center rounded-full border border-border transition-colors duration-300 text-text-gray hover:bg-primary hover:text-white"
      >
        <FaFacebookF size={15} />
      </a>

      {/* WhatsApp */}
      <a
        href={whatsappUrl}
        target="_blank"
        rel="noopener noreferrer"
        aria-label={`Share ${title} on WhatsApp`}
        className="flex h-9 w-9 items-center justify-center rounded-full border border-border transition-colors duration-300 text-text-gray hover:bg-primary hover:text-white"
      >
        <FaWhatsapp size={17} />
      </a>

      {/* Copy link */}
      <button
        type="button"
        onClick={copyLink}
        aria-label="Copy product link"
        className="flex h-9 w-9 items-center justify-center rounded-full border border-border transition-colors duration-300 text-text-gray hover:bg-primary hover:text-white"
      >
        {copied ? <FiCheck size={17} /> : <FiCopy size={17} />}
      </button>

      <span className="sr-only" aria-live="polite">
        {copied ? "Product link copied" : ""}
      </span>
    </div>
  );
};

export default SocialShare;