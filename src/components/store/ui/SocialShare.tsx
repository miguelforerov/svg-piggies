import { useState, type ReactNode } from "react";
import { FaFacebookF, FaPinterestP, FaWhatsapp } from "react-icons/fa";
import { FiCheck, FiCopy } from "react-icons/fi";
import { Button } from "@/components/ui/button";

const shareActionClassName =
  "flex h-8 w-8 items-center justify-center rounded-md transition-colors duration-300 text-primary hover:bg-primary hover:text-white";

interface SocialShareProps {
  title: string;
  url?: string;
  imageUrl?: string;
}

interface ShareLinkProps {
  href: string;
  ariaLabel: string;
  children: ReactNode;
}

const Link = ({ href, ariaLabel, children }: ShareLinkProps) => (
  <a
    href={href}
    target="_blank"
    rel="noopener noreferrer"
    aria-label={ariaLabel}
    className={shareActionClassName}
  >
    {children}
  </a>
);

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

  const pinterestUrl = `https://www.pinterest.com/pin/create/button/?url=${encodedUrl}&media=${encodedImageUrl}&description=${encodedTitle}`;
  const facebookUrl = `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`;
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
    <div className="flex items-center gap-4">
      <div className="text-sm font-medium text-primary">Share</div>

      <Link
        href={pinterestUrl}
        ariaLabel={`Share ${title} on Pinterest`}
      >
        <FaPinterestP size={16} />
      </Link>

      <Link
        href={facebookUrl}
        ariaLabel={`Share ${title} on Facebook`}
      >
        <FaFacebookF size={16} />
      </Link>

      <Link
        href={whatsappUrl}
        ariaLabel={`Share ${title} on WhatsApp`}
      >
        <FaWhatsapp size={16} />
      </Link>

      <Button
        variant="ghost"
        onClick={copyLink}
        aria-label="Copy product link"
        className={shareActionClassName}
      >
        {copied ? <FiCheck size={16} /> : <FiCopy size={16} />}
      </Button>

      <span className="sr-only" aria-live="polite">
        {copied ? "Product link copied" : ""}
      </span>
    </div>
  );
};

export default SocialShare;
