import type { SVGProps } from "react";

interface TrashProps extends SVGProps<SVGSVGElement> {}

const Trash = ({...props }: TrashProps) => (
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
      d="M9.09091 4.72727V5.45455H5.45455C5.05289 5.45455 4.72727 5.78016 4.72727 6.18182V6.90909C4.72727 7.31075 5.05289 7.63637 5.45455 7.63637H18.5455C18.9471 7.63637 19.2727 7.31075 19.2727 6.90909V6.18182C19.2727 5.78016 18.9471 5.45455 18.5455 5.45455H14.9091V4.72727C14.9091 4.32561 14.5835 4 14.1818 4H9.81818C9.41653 4 9.09091 4.32561 9.09091 4.72727Z"
    />
    <path
      fill="currentColor"
      d="M6.12589 8.72726H17.874L17.1887 17.9793C17.1042 19.1187 16.1553 20 15.0128 20H8.98709C7.84463 20 6.89562 19.1187 6.81123 17.9793L6.12589 8.72726Z"
    />
  </svg>
);

export default Trash;
