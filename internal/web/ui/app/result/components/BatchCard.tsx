import Link from 'next/link';

interface BatchCardProps {
  emoji: string;
  title: string;
  description: string;
  href: string;
  buttonText: string;
  buttonColor: string;
  hoverColor: string;
  subtitle: string;
}

export default function BatchCard({
  emoji,
  title,
  description,
  href,
  buttonText,
  buttonColor,
  hoverColor,
  subtitle,
}: BatchCardProps) {
  return (
    <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-2">
      <div className="text-center">
        <div className="text-6xl mb-6">{emoji}</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-4">{title}</h2>
        <p className="text-gray-600 mb-6">{description}</p>
        <div className="space-y-3">
          <Link
            href={href}
            className={`block w-full px-6 py-3 bg-gradient-to-r ${buttonColor} text-white rounded-xl hover:${hoverColor} transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium`}
          >
            {buttonText}
          </Link>
          <div className="text-sm text-gray-500">{subtitle}</div>
        </div>
      </div>
    </div>
  );
}