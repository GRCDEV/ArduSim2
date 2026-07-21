import AppLayout from "./components/layouts/app-layout";
import { useNavigation } from "./hooks/useNavigation";

export default function App() {
  const current = useNavigation((e) => e.current);

  return (
    <AppLayout>
      <current.page />
    </AppLayout>
  );
}
