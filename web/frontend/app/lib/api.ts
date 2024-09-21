const API_HOST_URL = process.env.NEXT_PUBLIC_API_HOST_URL || 'http://localhost:3000'

interface HerokuMigrationRequest {
    source: string;
    destination: string;
    herokuApiKey: string;
}

interface MigrationResponse {
    downloadUrl: string;
    // Add any other properties that the API response includes
}

export async function migratePaas(params: HerokuMigrationRequest): Promise<MigrationResponse> {
    const response = await fetch(`${API_HOST_URL}/api/migrate/${params.source}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(params),
    })

    if (!response.ok) {
        // if the response is a json object, we can parse it and throw a more specific error
        if (response.headers.get('Content-Type')?.includes('application/json')) {
            const error = await response.json()
            throw new Error(error)
        }

        throw new Error("An unexpected error occurred")
    }

    return await response.json()
}