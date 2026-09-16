import React from "react";

interface ButtonProps {
  label: string;
  link: string;
  style?: string;
  rel?: "follow" | "nofollow";
  newTab?: boolean;
}

const Button = ({
  label,
  link,
  style,
  rel,
  newTab = false,
}: ButtonProps) => {
  const relValue = [
    newTab ? "noopener" : null,
    newTab ? "noreferrer" : null,
    rel === "nofollow" ? "nofollow" : null,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <a
      href={link}
      target={newTab ? "_blank" : undefined}
      rel={relValue || undefined}
      className={`inline-block rounded-md bg-primary px-5 py-2 font-semibold text-white transition hover:bg-primary/80 hover:no-underline dark:hover:text-black ${style ?? ""}`}
    >
      {label}
    </a>
  );
};

export default Button;