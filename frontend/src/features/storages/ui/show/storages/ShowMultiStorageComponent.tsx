import { useEffect, useState } from 'react';

import { type Storage, storageApi } from '../../../../../entity/storages';

interface Props {
  storage: Storage;
}

export function ShowMultiStorageComponent({ storage }: Props) {
  const [primaryStorage, setPrimaryStorage] = useState<Storage | null>(null);
  const [secondaryStorage, setSecondaryStorage] = useState<Storage | null>(null);

  useEffect(() => {
    const fetchStorages = async () => {
      if (storage?.multiStorage?.primaryId) {
        try {
          const primary = await storageApi.getStorage(storage.multiStorage.primaryId);
          setPrimaryStorage(primary);
        } catch (e) {
          console.error('Failed to fetch primary storage:', e);
        }
      }

      if (storage?.multiStorage?.secondaryId) {
        try {
          const secondary = await storageApi.getStorage(storage.multiStorage.secondaryId);
          setSecondaryStorage(secondary);
        } catch (e) {
          console.error('Failed to fetch secondary storage:', e);
        }
      }
    };

    fetchStorages();
  }, [storage?.multiStorage?.primaryId, storage?.multiStorage?.secondaryId]);

  return (
    <>
      <div className="mb-1 flex items-center">
        <div className="min-w-[110px]">Primary</div>
        {primaryStorage?.name || storage?.multiStorage?.primaryId || '-'}
      </div>

      <div className="mb-1 flex items-center">
        <div className="min-w-[110px]">Secondary</div>
        {secondaryStorage?.name || storage?.multiStorage?.secondaryId || '-'}
      </div>
    </>
  );
}
