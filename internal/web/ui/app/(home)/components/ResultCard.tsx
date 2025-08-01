import InfoCard from '../../../components/InfoCard';

export default function ResultCard() {
  return (
    <InfoCard
      icon="📊"
      iconColor="blue-600"
      title="결과 조회"
      description="업로드된 파일의 결과와 다운로드 링크를 확인하세요."
      buttonText="샘플 결과 보기"
      buttonHref="/result/sample1"
      buttonColor="blue-600"
      footerText="샘플 ID: sample1, sample2, sample3"
    />
  );
}
