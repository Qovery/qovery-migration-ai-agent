const API_HOST_URL = process.env.NEXT_PUBLIC_API_HOST_URL || 'http://localhost:3000'

interface MigratePaasParams {
    source: string;
    destination: string;
    apiKey: string;
}

interface MigrationResponse {
    downloadUrl: string;
    // Add any other properties that the API response includes
}

export async function migratePaas({source, destination, apiKey}: MigratePaasParams): Promise<MigrationResponse> {
    const response = await fetch(`${API_HOST_URL}/api/migrate/${source}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({destination, apiKey}),
    })

    if (!response.ok) {
        throw new Error('Migration request failed')
    }

    return await response.json()
}