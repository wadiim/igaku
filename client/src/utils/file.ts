// NOTE: Does not work in Firefox.
export async function getNewCsvFileHandle(): Promise<FileSystemFileHandle> {
  const options = {
    types: [
      {
        description: 'CSV Files',
        accept: {
          'text/csv': ['.csv'] as ['.csv'],
        },
      },
    ],
  };
  if (!('showOpenFilePicker' in window)) {
    throw new Error("File Picker not supported by the browser");
  } else {
    const handle = await window.showSaveFilePicker(options);
    return handle;
  }
}
