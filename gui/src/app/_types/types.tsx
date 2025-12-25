
export interface Note {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    DeletedAt: string | null;
    Content: string;
    Highlight: boolean;
    Private: boolean;
    Frequency: number;
}