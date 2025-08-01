import InfoCard from './InfoCard';

export default function CliUploadCard() {
  return (
    <InfoCard
      icon="💻"
      iconColor="green-600"
      title="CLI 업로드"
      description="명령줄에서 고급 기능과 함께 빠르게 파일을 업로드하세요."
      buttonText="CLI 사용법 보기"
      buttonHref="/upload/cli"
      buttonColor="green-600"
    />
  );
}
