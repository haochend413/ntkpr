// noteTable - Server Component for fetching and displaying notes

import fs from 'fs';
import os from 'os';
import path from 'path';
import YAML from 'yaml';
import {Note} from "../_types/types";

// Force dynamic rendering (no static generation)
export const dynamic = 'force-dynamic';


function expandPath(p: string): string {
    if (p.includes('%APPDATA%')) {
        const appData = process.env.APPDATA;
        if (!appData) throw new Error('APPDATA not set');
        p = p.replace(/%APPDATA%/g, appData);
    }
    if (p.startsWith('~')) {
        p = path.join(os.homedir(), p.slice(1));
    }
    return p;
}

function getConfigPath(): string {
    const possiblePaths = [
        '~/Library/Application Support/ntkpr/config.yaml',
        '~/.config/ntkpr/config.yaml',
        '%APPDATA%\\ntkpr\\config.yaml',
    ];

    for (const configPath of possiblePaths) {
        try {
            const expanded = expandPath(configPath);
            console.log('[Server] Checking config path:', expanded);
            if (fs.existsSync(expanded)) {
                console.log('[Server] ✓ Found config at:', expanded);
                return expanded;
            }
        } catch (err) {
            console.error('[Server] Error checking path:', configPath, err);
            continue;
        }
    }
    throw new Error('Config file not found in any standard location');
}

function fetchNotes(): Note[] {
    console.log('[Server] Starting fetchNotes()');
    
    try {
        const configPath = getConfigPath();
        const file = fs.readFileSync(configPath, "utf8");
        console.log('[Server] Config file loaded, length:', file.length);
        
        const data = YAML.parse(file);
        console.log('[Server] Parsed config:', data);
        
        const dataPath = data.datafilepath || data.DataFilePath;
        if (!dataPath) {
            throw new Error('datafilepath not found in config');
        }
        console.log('[Server] Data path from config:', dataPath);
        
        const notesDataPath = path.join(dataPath, "notes.json");
        console.log('[Server] Full notes JSON path:', notesDataPath);
        
        if (!fs.existsSync(notesDataPath)) {
            throw new Error(`JSON file not found at: ${notesDataPath}`);
        }

        const jsonData = fs.readFileSync(notesDataPath, 'utf8');
        const notes: Note[] = JSON.parse(jsonData);
        console.log('[Server] Retrieved', notes.length, 'notes from JSON');
        
        return notes;
    } catch (error) {
        console.error('[Server] ✗ Error fetching notes:', error);
        return [];
    }
}

// Server Component - runs on server, can use Node.js APIs
export default function NoteTable() {
    console.log('[Server] NoteTable component rendering');
    const notes = fetchNotes();
    console.log('[Server] Rendering', notes.length, 'notes in table');
    
    if (notes.length === 0) {
        return (
            <div className="content">
                <p>No notes found. Check server console for errors.</p>
            </div>
        );
    }
    
    return (
        <div className="content">
        <table className="list">
          <thead>
            <tr>
              <th>ID</th>
              <th>Preview</th>
              <th>CreatedAt</th>
              <th>Latest UpdatedAt</th>
              <th>Frequency</th>
              <th>Highlighted</th>
            </tr>
          </thead>
          <tbody id="rows">
            {notes.map((note) => (
              <tr key={note.ID}>
                <td>{note.ID}</td>
                <td>{note.Content.substring(0, 50)}{note.Content.length > 50 ? '...' : ''}</td>
                <td>{new Date(note.CreatedAt).toLocaleString('en-US', { hour12: false })}</td>
                <td>{new Date(note.UpdatedAt).toLocaleString('en-US', { hour12: false })}</td>
                <td>{note.Frequency}</td>
                <td>{note.Highlight ? 'H' : ''}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
}