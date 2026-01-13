import { Select, Alert } from 'antd';
import { useEffect, useState } from 'react';

import { type Storage, StorageType, storageApi } from '../../../../../entity/storages';

interface Props {
  storage: Storage;
  setStorage: (storage: Storage) => void;
  setUnsaved: () => void;
}

export function EditMultiStorageComponent({ storage, setStorage, setUnsaved }: Props) {
  const [availableStorages, setAvailableStorages] = useState<Storage[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchStorages = async () => {
      setIsLoading(true);
      try {
        const storages = await storageApi.getStorages(storage.workspaceId);
        // Filter out MULTI storages and the current storage being edited
        const filtered = storages.filter(
          (s) => s.type !== StorageType.MULTI && s.id !== storage.id,
        );
        setAvailableStorages(filtered);
      } catch (e) {
        console.error('Failed to fetch storages:', e);
      }
      setIsLoading(false);
    };

    fetchStorages();
  }, [storage.workspaceId, storage.id]);

  const storageOptions = availableStorages.map((s) => ({
    label: s.name,
    value: s.id,
  }));

  return (
    <>
      <Alert
        type="info"
        message="Multi Storage saves backups to two different storages simultaneously. Both storages must succeed for the backup to be successful."
      />

      <div className="mt-5" />

      <div className="mb-1 flex w-full flex-col items-start sm:flex-row sm:items-center">
        <div className="mb-1 min-w-[110px] sm:mb-0">Primary</div>
        <Select
          value={storage?.multiStorage?.primaryId || undefined}
          options={storageOptions.filter((o) => o.value !== storage?.multiStorage?.secondaryId)}
          onChange={(value) => {
            if (!storage?.multiStorage) return;

            setStorage({
              ...storage,
              multiStorage: {
                ...storage.multiStorage,
                primaryId: value,
              },
            });
            setUnsaved();
          }}
          loading={isLoading}
          size="small"
          className="w-full max-w-[250px]"
          placeholder="Select primary storage"
        />
      </div>

      <div className="mb-1 flex w-full flex-col items-start sm:flex-row sm:items-center">
        <div className="mb-1 min-w-[110px] sm:mb-0">Secondary</div>
        <Select
          value={storage?.multiStorage?.secondaryId || undefined}
          options={storageOptions.filter((o) => o.value !== storage?.multiStorage?.primaryId)}
          onChange={(value) => {
            if (!storage?.multiStorage) return;

            setStorage({
              ...storage,
              multiStorage: {
                ...storage.multiStorage,
                secondaryId: value,
              },
            });
            setUnsaved();
          }}
          loading={isLoading}
          size="small"
          className="w-full max-w-[250px]"
          placeholder="Select secondary storage"
        />
      </div>

      {availableStorages.length < 2 && !isLoading && (
        <Alert
          type="warning"
          message="You need at least two non-multi storages to create a Multi Storage."
          className="mt-4"
        />
      )}

      <div className="mb-5" />
    </>
  );
}
