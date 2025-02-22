
interface CardProps {
  overline: string;
  value: string | number;
  description?: string;
  isLarge?: boolean;
}

const Card = ({ overline, value = 'Keine Angabe', description = '', isLarge = false }: CardProps) => {
  return (
    <div className="h-full space-y-3 bg-dark-50 rounded-xl p-6">
      <h2 className="text-sm text-dark-700 font-medium">{overline}</h2>
      <p className={`font-bold ${isLarge ? 'text-3xl' : 'text-xl'}`}>{value}</p>
      {description && <p className="text-sm">{description}</p>}
    </div>
  );
}

export default Card;

