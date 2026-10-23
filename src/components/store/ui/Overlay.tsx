interface OverlayProps {
  isOpen: boolean;
  onClick?: () => void;
}

function Overlay({ isOpen, onClick }: OverlayProps) {
  return (
    <div
      className={`fixed inset-0 z-40 bg-black transition-opacity backdrop-blur-xs ${
        isOpen
          ? "pointer-events-auto opacity-80"
          : "pointer-events-none opacity-0"
      }`}
      onClick={onClick}
      aria-hidden="true"
    />
  );
};

export default Overlay