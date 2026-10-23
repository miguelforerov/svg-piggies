import type { SVGProps } from "react";

interface MinusProps extends SVGProps<SVGSVGElement> {}

const Minus = ({ ...props }: MinusProps) => (
  <svg
    width="100%"
    height="100%"
    fill="none"
    aria-hidden="true"
    viewBox="0 0 24 24"
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <path
      fill="currentColor"
      fill-rule="evenodd"
      clip-rule="evenodd"
      d="M17 13C17.5523 13 18 12.5523 18 12C18 11.4477 17.5523 11 17 11C10.5938 11 12.7931 11 7 11C6.44772 11 6 11.4477 6 12C6 12.5523 6.44772 13 7 13C12.7758 13 10.6153 13 17 13Z"
    />
  </svg>
);

export default Minus;
