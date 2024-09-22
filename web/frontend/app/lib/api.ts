const API_HOST_URL = process.env.NEXT_PUBLIC_API_HOST_URL || 'http://localhost:3000'

interface HerokuMigrationRequest {
    source: string;
    destination: string;
    herokuApiKey: string;
}

interface MigrationResponse {
    blob: Blob;
    filename: string;
}

export async function generateMigrationFiles(params: HerokuMigrationRequest): Promise<MigrationResponse> {
    const response = await fetch(`${API_HOST_URL}/api/migrate/${params.source}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(params),
    })

    if (!response.ok) {
        let errorMessage = "An unexpected error occurred";
        if (response.headers.get('Content-Type')?.includes('application/json')) {
            const errorData = await response.json();
            errorMessage = errorData.message || errorMessage;
        }
        throw new Error(errorMessage);
    }

    // Check if the response is a zip file
    const contentType = response.headers.get('Content-Type');
    if (contentType === 'application/zip') {
        const blob = await response.blob();
        const filename = getFilenameFromContentDisposition(response.headers.get('Content-Disposition')) || 'download.zip';
        return {blob, filename};
    } else {
        throw new Error("Unexpected response format");
    }
}

function getFilenameFromContentDisposition(contentDisposition: string | null): string | null {
    if (!contentDisposition) return null;
    const filenameMatch = contentDisposition.match(/filename="?(.+)"?/i);
    return filenameMatch ? filenameMatch[1] : null;
}