import { RetrieveSecret } from "@/components/retrieve-secret";

export default async function SecretPage({ params }: { params: Promise<{ secretId: string }> }) {
  const { secretId } = await params;
  return <RetrieveSecret secretId={secretId} />;
}
