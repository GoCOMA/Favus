import Link from 'next/link';

type InfoCardProps = {
  icon: string;
  iconColor: string;
  title: string;
  description: string;
  buttonText: string;
  buttonHref: string;
  buttonColor: string;
  footerText?: string;
};

export default function InfoCard({
  icon,
  iconColor,
  title,
  description,
  buttonText,
  buttonHref,
  buttonColor,
  footerText,
}: InfoCardProps) {
  return (
    <div className="bg-white rounded-lg shadow-sm p-8 hover:shadow-md transition-shadow mb-8">
      <div className="text-center">
        <div className={`text-${iconColor} text-5xl mb-4`}>{icon}</div>
        <h2 className="text-2xl font-semibold text-gray-900 mb-4">{title}</h2>
        <p className="text-gray-600 mb-6">{description}</p>
        <Link
          href={buttonHref}
          className={`inline-flex items-center px-6 py-3 bg-${buttonColor} text-white rounded-lg hover:bg-${buttonColor}-700 transition-colors`}
        >
          {buttonText}
        </Link>
        {footerText && (
          <p className="text-sm text-gray-500 mt-3">{footerText}</p>
        )}
      </div>
    </div>
  );
}
