import { useId, useState } from 'react';
import {
  createAdminUser,
  disableAdminUser,
  enableAdminUser,
  errorMessage,
  fetchUserCapabilities,
  grantUserCapability,
  issueMagicToken,
  revokeUserCapability
} from './api';
import type { AdminUser, UserCapabilitiesResponse } from './types';

type ToastRequest = {
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
};

type AdminViewProps = {
  users: AdminUser[];
  error: string | null;
  onRefresh: () => Promise<void>;
  notify: (toast: ToastRequest) => void;
  formatTimestamp: (value: string) => string;
};

export function AdminView({ users, error, onRefresh, notify, formatTimestamp }: AdminViewProps) {
  const [username, setUsername] = useState('');
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null);
  const [capabilities, setCapabilities] = useState<UserCapabilitiesResponse | null>(null);
  const [selectedCapability, setSelectedCapability] = useState('');
  const [issuedToken, setIssuedToken] = useState<{ token: string; expiresAt: string; username: string } | null>(null);
  const grantOptions = capabilities?.available.filter((capability) => !capabilities.capabilities.includes(capability)) ?? [];

  async function createUser() {
    try {
      const user = await createAdminUser(username.trim());
      setUsername('');
      await onRefresh();
      notify({ type: 'success', title: 'User created', message: `${user.username} is ready. Grant only the needed capabilities, then issue a one-time magic token.` });
    } catch (adminError: unknown) {
      notify({ type: 'error', title: 'User creation failed', message: errorMessage(adminError, 'Koji could not create the user. Check the username and your permission.') });
    }
  }

  async function toggleUser(user: AdminUser) {
    try {
      const updated = user.disabled ? await enableAdminUser(user.id) : await disableAdminUser(user.id);
      await onRefresh();
      notify(userToggleToast(updated));
    } catch (adminError: unknown) {
      notify({ type: 'error', title: 'User update failed', message: errorMessage(adminError, 'Koji could not update the user. Check whether this would remove the final administrator.') });
    }
  }

  async function manageCapabilities(user: AdminUser) {
    try {
      const response = await fetchUserCapabilities(user.id);
      setSelectedUser(user);
      setCapabilities(response);
      setSelectedCapability(firstGrantOption(response));
    } catch (adminError: unknown) {
      notify({ type: 'error', title: 'Capabilities unavailable', message: errorMessage(adminError, 'Koji could not load capabilities for this user.') });
    }
  }

  async function grantCapability() {
    if (!selectedUser || selectedCapability === '') {
      return;
    }
    try {
      const assigned = await grantUserCapability(selectedUser.id, selectedCapability);
      setCapabilities((current) => replaceAssignedCapabilities(current, assigned));
      notify({ type: 'success', title: 'Capability granted', message: `${selectedCapability} was granted to ${selectedUser.username}. The change is active for future requests.` });
    } catch (adminError: unknown) {
      notify({ type: 'error', title: 'Grant failed', message: errorMessage(adminError, 'Koji could not grant that capability.') });
    }
  }

  async function revokeCapability(capability: string) {
    if (!selectedUser) {
      return;
    }
    try {
      const assigned = await revokeUserCapability(selectedUser.id, capability);
      setCapabilities((current) => replaceAssignedCapabilities(current, assigned));
      notify({ type: 'success', title: 'Capability revoked', message: `${capability} was revoked from ${selectedUser.username}.` });
    } catch (adminError: unknown) {
      notify({ type: 'error', title: 'Revoke failed', message: errorMessage(adminError, 'Koji could not revoke that capability. The final identity administrator is protected.') });
    }
  }

  async function issueToken(user: AdminUser) {
    try {
      const token = await issueMagicToken(user.id);
      setIssuedToken({ ...token, username: user.username });
      notify({ type: 'warning', title: 'Magic token issued', message: 'Copy this token now and deliver it through an approved operator channel. Koji will not show it again.' });
    } catch (adminError: unknown) {
      notify({ type: 'error', title: 'Token issue failed', message: errorMessage(adminError, 'Koji could not issue a token. Enabled users only can receive tokens.') });
    }
  }

  return (
    <div className="page-stack">
      {error && <InlineError message={error} />}
      <PermissionNotice message="Identity administration requires identity.users.manage. Managed users sign in with one-time magic tokens instead of passwords." />
      <CreateUserPanel username={username} onUsernameChange={setUsername} onCreate={createUser} />
      {issuedToken && <MagicTokenPanel issuedToken={issuedToken} formatTimestamp={formatTimestamp} />}
      <UsersPanel users={users} formatTimestamp={formatTimestamp} onCapabilities={manageCapabilities} onIssueToken={issueToken} onToggle={toggleUser} />
      {selectedUser && capabilities && (
        <CapabilitiesPanel
          user={selectedUser}
          capabilities={capabilities.capabilities}
          grantOptions={grantOptions}
          selectedCapability={selectedCapability}
          onSelectedCapabilityChange={setSelectedCapability}
          onGrant={grantCapability}
          onRevoke={revokeCapability}
        />
      )}
    </div>
  );
}

function CreateUserPanel({
  username,
  onUsernameChange,
  onCreate
}: {
  username: string;
  onUsernameChange: (value: string) => void;
  onCreate: () => void;
}) {
  return (
    <section className="activity-panel">
      <div className="card-heading">
        <h3>Create Managed User</h3>
        <Tooltip text="Managed users do not have passwords. Issue a one-time magic token after creating the user." />
      </div>
      <div className="admin-inline-form">
        <input value={username} onChange={(event) => onUsernameChange(event.target.value)} placeholder="Username" aria-label="New managed username" />
        <button type="button" onClick={onCreate} disabled={username.trim() === ''}>Create user</button>
      </div>
    </section>
  );
}

function MagicTokenPanel({
  issuedToken,
  formatTimestamp
}: {
  issuedToken: { token: string; expiresAt: string; username: string };
  formatTimestamp: (value: string) => string;
}) {
  return (
    <section className="activity-panel magic-token-panel" aria-label="Issued magic token">
      <div className="card-heading">
        <h3>Magic Token For {issuedToken.username}</h3>
        <Tooltip text="Copy this token now. Koji stores only a hash and cannot show the raw token again." />
      </div>
      <code>{issuedToken.token}</code>
      <p>Copy this token now. It expires {formatTimestamp(issuedToken.expiresAt)} and cannot be shown again.</p>
    </section>
  );
}

function UsersPanel({
  users,
  formatTimestamp,
  onCapabilities,
  onIssueToken,
  onToggle
}: {
  users: AdminUser[];
  formatTimestamp: (value: string) => string;
  onCapabilities: (user: AdminUser) => void;
  onIssueToken: (user: AdminUser) => void;
  onToggle: (user: AdminUser) => void;
}) {
  return (
    <section className="activity-panel">
      <div className="card-heading">
        <h3>Users</h3>
        <Tooltip text="The final Super Admin and final identity manager are protected from lockout." />
      </div>
      <div className="table-wrapper">
        <table className="activity-table">
          <thead>
            <tr>
              <th>Username</th>
              <th>Type</th>
              <th>Status</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <tr key={user.id}>
                <td>{user.username}</td>
                <td>{user.isSuperAdmin ? 'Super Admin' : 'Managed'}</td>
                <td><StatusBadge status={user.disabled ? 'fail' : 'ok'} label={user.disabled ? 'disabled' : 'active'} /></td>
                <td>{formatTimestamp(user.createdAt)}</td>
                <td>
                  <div className="job-decision-controls">
                    <button type="button" onClick={() => onCapabilities(user)}>Capabilities</button>
                    <button type="button" onClick={() => onIssueToken(user)} disabled={user.disabled} title={user.disabled ? 'Enable this user before issuing a token.' : 'Issue a one-time sign-in token.'}>Issue token</button>
                    <button type="button" onClick={() => onToggle(user)}>{user.disabled ? 'Enable' : 'Disable'}</button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function CapabilitiesPanel({
  user,
  capabilities,
  grantOptions,
  selectedCapability,
  onSelectedCapabilityChange,
  onGrant,
  onRevoke
}: {
  user: AdminUser;
  capabilities: string[];
  grantOptions: string[];
  selectedCapability: string;
  onSelectedCapabilityChange: (capability: string) => void;
  onGrant: () => void;
  onRevoke: (capability: string) => void;
}) {
  return (
    <section className="activity-panel">
      <div className="card-heading">
        <h3>Capabilities For {user.username}</h3>
        <Tooltip text="Capabilities decide what an authenticated user may do. Revoking the final identity manager is blocked." />
      </div>
      <div className="admin-inline-form">
        <select value={selectedCapability} onChange={(event) => onSelectedCapabilityChange(event.target.value)} aria-label="Capability to grant">
          <option value="">Select capability</option>
          {grantOptions.map((capability) => <option key={capability} value={capability}>{capability}</option>)}
        </select>
        <button type="button" onClick={onGrant} disabled={selectedCapability === ''}>Grant</button>
      </div>
      <div className="capability-list">
        {capabilities.map((capability) => (
          <span key={capability} className="capability-chip">
            {capability}
            <button type="button" onClick={() => onRevoke(capability)} aria-label={`Revoke ${capability}`}>Revoke</button>
          </span>
        ))}
      </div>
    </section>
  );
}

function replaceAssignedCapabilities(current: UserCapabilitiesResponse | null, assigned: string[]): UserCapabilitiesResponse | null {
  return current ? { ...current, capabilities: assigned } : current;
}

function firstGrantOption(response: UserCapabilitiesResponse): string {
  return response.available.find((capability) => !response.capabilities.includes(capability)) ?? '';
}

function userToggleToast(user: AdminUser): ToastRequest {
  if (user.disabled) {
    return { type: 'success', title: 'User disabled', message: `${user.username} can no longer sign in.` };
  }
  return { type: 'success', title: 'User enabled', message: `${user.username} can sign in with a valid magic token.` };
}

function InlineError({ message }: { message: string }) {
  return <InlineNotice tone="error" message={message} />;
}

function PermissionNotice({ message }: { message: string }) {
  return <InlineNotice tone="info" message={message} />;
}

function InlineNotice({
  tone,
  message
}: {
  tone: 'error' | 'info' | 'warning';
  message: string;
}) {
  return (
    <div className={`inline-notice ${tone}`} role={tone === 'error' ? 'alert' : 'note'}>
      <strong>{noticePrefix(tone)}</strong>
      <span>{message}</span>
    </div>
  );
}

function StatusBadge({ status, label }: { status: 'ok' | 'fail'; label: string }) {
  return <span className={`status-badge ${status}`}>{status === 'ok' ? 'OK' : 'FAIL'} {label}</span>;
}

function Tooltip({ text }: { text: string }) {
  const id = useId();
  return (
    <span className="tooltip" tabIndex={0} aria-label="Help" aria-describedby={id}>
      ?
      <span id={id} className="tooltip-content" role="tooltip">{text}</span>
    </span>
  );
}

function noticePrefix(tone: 'error' | 'info' | 'warning'): string {
  if (tone === 'error') {
    return 'Error';
  }
  if (tone === 'warning') {
    return 'Notice';
  }
  return 'Info';
}
