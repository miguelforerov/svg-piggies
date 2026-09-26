import { BsCheckLg } from "react-icons/bs";

interface CheckboxProps {
  checked: boolean;
}

const Checkbox = ({ checked }: CheckboxProps) => (
  <span
    aria-hidden="true"
    className="flex h-4 w-4 items-center justify-center rounded-sm border border-border-primary"
  >
    {checked && <BsCheckLg size={16} />}
  </span>
);

export default Checkbox;
