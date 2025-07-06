// CLI 업로드 안내

export default function CliGuidePage() {
  return (
    <main className="p-8">
      <h1 className="text-2xl font-semibold">CLI 업로드 가이드</h1>
      <pre className="mt-4 bg-gray-100 p-4 rounded-md text-sm">
        {`$ go install github.com/favus/cli@latest
  $ favus upload ./yourfile.zip`}
      </pre>
    </main>
  );
}
