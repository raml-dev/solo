export interface StableKeyValueRow {
  id: string;
  key: string;
  value: string;
}

export function createStableId(): string {
  try {
    return crypto.randomUUID();
  } catch {
    return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
  }
}

export function mapRecordToRowsWithStableIds(
  values: Record<string, string> = {},
  previous: Pick<StableKeyValueRow, "id" | "key">[] = []
): StableKeyValueRow[] {
  const idsByKey = new Map<string, string[]>();

  for (const row of previous) {
    const trimmedKey = row.key.trim();
    if (!trimmedKey) continue;
    const bucket = idsByKey.get(trimmedKey) ?? [];
    bucket.push(row.id);
    idsByKey.set(trimmedKey, bucket);
  }

  const takeId = (key: string) => {
    const bucket = idsByKey.get(key);
    if (bucket && bucket.length > 0) {
      const id = bucket.shift();
      if (id) return id;
    }
    return createStableId();
  };

  return Object.entries(values).map(([key, value]) => ({
    id: takeId(key),
    key,
    value
  }));
}
