import React from 'react';

interface PrimaryButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, React.PropsWithChildren {
  isDanger?: boolean;
}

const PrimaryButton = ({ children, isDanger = false, ...props }: PrimaryButtonProps) => (
  <button
    {...props}
    className={`${isDanger ? 'bg-red' : 'bg-green-dark hover:bg-green-light'} text-white w-fit px-5 py-2 group flex gap-x-3 rounded-xl items-center transition-all ease-in-out duration-300 disabled:bg-dark-400 ${props.className}`}
  >
    {children}
  </button >
);

export default PrimaryButton;
