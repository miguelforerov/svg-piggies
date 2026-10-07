import type { SVGProps } from "react";

const Profile = (props: SVGProps<SVGSVGElement>) => (
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
      d="M12 10.5714C14.4853 10.5714 16.5 8.65264 16.5 6.28572C16.5 3.91879 14.4853 2 12 2C9.51472 2 7.5 3.91879 7.5 6.28572C7.5 8.65264 9.51472 10.5714 12 10.5714Z"
    />
    <path
      fill="currentColor"
      d="M21 17.7143C21 15.3473 18.9853 13.4286 16.5 13.4286H7.5C5.01472 13.4286 3 15.3473 3 17.7143V22H21V17.7143Z"
  />
  </svg>
);

export default Profile;
