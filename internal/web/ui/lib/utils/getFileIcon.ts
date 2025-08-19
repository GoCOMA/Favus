export const getFileIcon = (fileName: string) => {
  const extension = fileName.split('.').pop()?.toLowerCase();
  const map: Record<string, string> = {
    pdf: '📄',
    doc: '📝',
    docx: '📝',
    xls: '📊',
    xlsx: '📊',
    ppt: '📈',
    pptx: '📈',
    jpg: '🖼️',
    jpeg: '🖼️',
    png: '🖼️',
    gif: '🖼️',
    svg: '🖼️',
    mp4: '🎥',
    avi: '🎥',
    mov: '🎥',
    mp3: '🎵',
    wav: '🎵',
    flac: '🎵',
    zip: '📦',
    rar: '📦',
    '7z': '📦',
    txt: '📄',
  };
  return map[extension ?? ''] || '📁';
};
