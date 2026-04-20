import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { RenameInput } from '../components/RenameInput';

describe('RenameInput', () => {
  it('displays the value as a clickable heading', () => {
    render(<RenameInput value="My Droplet" onSave={vi.fn()} />);
    expect(screen.getByText('My Droplet')).toBeInTheDocument();
  });

  it('enters edit mode on click', async () => {
    render(<RenameInput value="My Droplet" onSave={vi.fn()} />);
    const heading = screen.getByText('My Droplet');
    fireEvent.click(heading);
    const input = screen.getByRole('textbox', { name: 'Rename droplet' });
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue('My Droplet');
  });

  it('saves on Enter key', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<RenameInput value="Old Title" onSave={onSave} />);
    const heading = screen.getByText('Old Title');
    fireEvent.click(heading);
    const input = screen.getByRole('textbox', { name: 'Rename droplet' });
    fireEvent.change(input, { target: { value: 'New Title' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onSave).toHaveBeenCalledWith('New Title');
  });

  it('cancels on Escape key', () => {
    render(<RenameInput value="My Droplet" onSave={vi.fn()} />);
    const heading = screen.getByText('My Droplet');
    fireEvent.click(heading);
    const input = screen.getByRole('textbox', { name: 'Rename droplet' });
    fireEvent.change(input, { target: { value: 'Changed' } });
    fireEvent.keyDown(input, { key: 'Escape' });
    expect(screen.getByText('My Droplet')).toBeInTheDocument();
  });

  it('does not call onSave if value unchanged', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<RenameInput value="Same Title" onSave={onSave} />);
    const heading = screen.getByText('Same Title');
    fireEvent.click(heading);
    const input = screen.getByRole('textbox', { name: 'Rename droplet' });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onSave).not.toHaveBeenCalled();
  });

  it('shows error message when rename fails', async () => {
    const onSave = vi.fn().mockRejectedValue(new Error('Rename failed'));
    render(<RenameInput value="Old Title" onSave={onSave} />);
    const heading = screen.getByText('Old Title');
    fireEvent.click(heading);
    const input = screen.getByRole('textbox', { name: 'Rename droplet' });
    fireEvent.change(input, { target: { value: 'New Title' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await screen.findByText('Rename failed');
  });
});