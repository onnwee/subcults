# Streaming Components

This directory contains React components for LiveKit audio streaming functionality.

## Components

### JoinStreamButton
Button component for joining a LiveKit audio room.

**Props:**
- `isConnected: boolean` - Whether currently connected to room
- `isConnecting: boolean` - Whether connection is in progress
- `onJoin: () => void` - Callback when join button is clicked
- `disabled?: boolean` - Whether button is disabled

**Usage:**
```tsx
<JoinStreamButton
  isConnected={isConnected}
  isConnecting={isConnecting}
  onJoin={handleJoin}
/>
```

### ParticipantList
Displays a list of participants in the audio room with their mute and speaking state.

**Props:**
- `participants: Participant[]` - List of remote participants
- `localParticipant: Participant | null` - Local participant (shown first with "You" label)

**Usage:**
```tsx
<ParticipantList
  participants={remoteParticipants}
  localParticipant={localParticipant}
/>
```

### AudioControls
Control panel for audio settings: mute/unmute, volume adjustment, and leaving the room.

**Props:**
- `isMuted: boolean` - Whether local microphone is muted
- `onToggleMute: () => void` - Callback to toggle mute state
- `onLeave: () => void` - Callback to leave the room
- `onVolumeChange: (volume: number) => void` - Callback when volume changes (0-100)
- `disabled?: boolean` - Whether controls are disabled

**Usage:**
```tsx
<AudioControls
  isMuted={isMuted}
  onToggleMute={toggleMute}
  onLeave={disconnect}
  onVolumeChange={setVolume}
/>
```

### ConnectionIndicator
Visual indicator showing connection quality based on LiveKit statistics.

**Props:**
- `quality: ConnectionQuality` - Quality level: 'excellent' | 'good' | 'poor' | 'unknown'
- `showLabel?: boolean` - Whether to show text label (default: true)

**Usage:**
```tsx
<ConnectionIndicator quality={connectionQuality} />
```

## Hook: useLiveAudio

Custom hook for managing LiveKit room connections.

**Parameters:**
- `roomName: string | null` - Name of the room to join
- `options?: UseLiveAudioOptions` - Optional configuration
  - `sceneId?: string` - Associated scene ID
  - `eventId?: string` - Associated event ID
  - `onError?: (error: Error) => void` - Error callback

**Returns:**
- `isConnected: boolean` - Connection state
- `isConnecting: boolean` - Connection in progress
- `participants: Participant[]` - Remote participants
- `localParticipant: Participant | null` - Local participant
- `connectionQuality: ConnectionQuality` - Current connection quality
- `error: string | null` - Error message if any
- `connect: () => Promise<void>` - Connect to room
- `disconnect: () => void` - Disconnect from room
- `toggleMute: () => Promise<void>` - Toggle microphone mute
- `setVolume: (volume: number) => void` - Set playback volume (0-100)

**Usage:**
```tsx
const {
  isConnected,
  isConnecting,
  participants,
  localParticipant,
  connectionQuality,
  error,
  connect,
  disconnect,
  toggleMute,
  setVolume,
} = useLiveAudio('room-name', {
  sceneId: 'scene-123',
  onError: (err) => console.error(err),
});
```

## Complete Example

See `StreamPage.tsx` for a complete implementation example.

```tsx
import { useLiveAudio } from '../hooks/useLiveAudio';
import {
  JoinStreamButton,
  ParticipantList,
  AudioControls,
  ConnectionIndicator,
} from '../components/streaming';

function StreamRoom({ roomName }: { roomName: string }) {
  const {
    isConnected,
    isConnecting,
    participants,
    localParticipant,
    connectionQuality,
    connect,
    disconnect,
    toggleMute,
    setVolume,
  } = useLiveAudio(roomName);

  return (
    <div>
      {!isConnected && (
        <JoinStreamButton
          isConnected={isConnected}
          isConnecting={isConnecting}
          onJoin={connect}
        />
      )}

      {isConnected && (
        <>
          <ConnectionIndicator quality={connectionQuality} />
          <AudioControls
            isMuted={localParticipant?.isMuted ?? true}
            onToggleMute={toggleMute}
            onLeave={disconnect}
            onVolumeChange={setVolume}
          />
          <ParticipantList
            participants={participants}
            localParticipant={localParticipant}
          />
        </>
      )}
    </div>
  );
}
```

## Testing

All components include comprehensive tests. Run with:

```bash
npm test -- streaming
```

## Known Limitations

- **Token Refresh**: LiveKit 2.x requires reconnection for token refresh. The hook schedules token refresh 30s before expiry but doesn't implement seamless reconnection yet. Users will need to manually rejoin after token expiry.

- **Volume Control**: Volume control uses a best-effort approach and may not work in all LiveKit versions. The implementation gracefully handles missing volume APIs.

## Security Considerations

- Never display hidden participant metadata in the UI
- LiveKit tokens are fetched from the backend API endpoint `/livekit/token`
- Tokens are short-lived (5 minutes by default)
- All token requests require authentication
