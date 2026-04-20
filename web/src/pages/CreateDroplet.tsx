import { useNavigate } from 'react-router-dom';
import { CreateDropletForm } from '../components/CreateDropletForm';
import type { Droplet } from '../api/types';

export function CreateDroplet() {
  const navigate = useNavigate();

  const handleSuccess = (droplet: Droplet) => {
    navigate(`/app/droplets/${droplet.id}`);
  };

  const handleCancel = () => {
    navigate('/app/droplets');
  };

  return (
    <div className="p-4 md:p-6 max-w-2xl mx-auto">
      <h1 className="text-xl font-mono font-bold text-cistern-fg mb-6">Create Droplet</h1>
      <div className="bg-cistern-surface border border-cistern-border rounded-lg p-6">
        <CreateDropletForm onSuccess={handleSuccess} onCancel={handleCancel} />
      </div>
    </div>
  );
}