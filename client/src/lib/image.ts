interface CompressImageOptions {
  maxSize?: number;
  quality?: number;
  skipIfSmallerThan?: number;
}

export async function compressImage(
  file: File,
  options: CompressImageOptions = {}
): Promise<File> {
  const {
    maxSize = 1920,
    quality = 0.8,
    skipIfSmallerThan = 1024 * 1024,
  } = options;

  if (file.size <= skipIfSmallerThan) return file;

  let objectUrl: string | null = null;
  try {
    objectUrl = URL.createObjectURL(file);
    const img = document.createElement("img");
    await new Promise<void>((resolve, reject) => {
      img.onload = () => resolve();
      img.onerror = () => reject(new Error("failed to load image"));
      img.src = objectUrl as string;
    });

    const { naturalWidth, naturalHeight } = img;
    if (!naturalWidth || !naturalHeight) return file;

    const scale = Math.min(1, maxSize / Math.max(naturalWidth, naturalHeight));
    const targetWidth = Math.round(naturalWidth * scale);
    const targetHeight = Math.round(naturalHeight * scale);

    const canvas = document.createElement("canvas");
    canvas.width = targetWidth;
    canvas.height = targetHeight;
    const ctx = canvas.getContext("2d");
    if (!ctx) return file;
    ctx.drawImage(img, 0, 0, targetWidth, targetHeight);

    const blob = await new Promise<Blob | null>((resolve) =>
      canvas.toBlob(resolve, "image/jpeg", quality)
    );
    if (!blob || blob.size >= file.size) return file;

    const baseName = file.name.replace(/\.\w+$/, "");
    return new File([blob], `${baseName}.jpg`, { type: "image/jpeg" });
  } catch (err) {
    console.error("画像の圧縮に失敗したため、元のファイルを使用します:", err);
    return file;
  } finally {
    if (objectUrl) URL.revokeObjectURL(objectUrl);
  }
}
