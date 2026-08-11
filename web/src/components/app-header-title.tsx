import { Link } from '@tanstack/react-router';
import { Brand } from './brand';

export default function AppHeaderTitle() {
  return (
    <Link to="/">
      <Brand />
    </Link>
  );
}
