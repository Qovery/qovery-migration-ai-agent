const API_HOST_URL = process.env.NEXT_PUBLIC_API_HOST_URL || 'http://localhost:3000'

interface MigrationRequest {
    source: string;
    destination: string;
    herokuApiKey?: string;
    cleverCloudToken?: string;
    cleverCloudSecret?: string;
}

interface MigrationResponse {
    blob: Blob;
    filename: string;
}

export async function generateMigrationFiles(migrationRequest: MigrationRequest): Promise<MigrationResponse> {
    const response = await fetch(`${API_HOST_URL}/api/migrate/${migrationRequest.source}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(migrationRequest),
    });

    if (!response.ok) {
        const errorMessage = await getErrorMessage(response);
        throw new Error(errorMessage);
    }

    const contentType = response.headers.get('Content-Type');
    if (contentType !== 'application/zip') {
        throw new Error("Unexpected response format");
    }

    const blob = await response.blob();
    const filename = getFilenameFromContentDisposition(response.headers.get('Content-Disposition')) || 'migration.zip';
    return {blob, filename};
}

async function getErrorMessage(response: Response): Promise<string> {
    const contentType = response.headers.get('Content-Type');
    if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        return errorData.message || "An unexpected error occurred";
    }
    return "An unexpected error occurred";
}

function getFilenameFromContentDisposition(contentDisposition: string | null): string | null {
    if (!contentDisposition) return null;
    const filenameMatch = contentDisposition.match(/filename="?(.+)"?/i);
    return filenameMatch ? filenameMatch[1] : null;
}