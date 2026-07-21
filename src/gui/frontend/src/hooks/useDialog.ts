import { create } from "zustand";
import type { ConfirmDialogProps } from "../components/dialogs/confirm-dialog";

interface State {
  dialogProps?: ConfirmDialogProps;
  show({ buttons, title, text }: ConfirmDialogProps): Promise<number>;
}

export const useDialog = create<State>((set) => ({
  async show({ buttons, title, text, onExit }: ConfirmDialogProps) {
    return new Promise((resolve, reject) => {
      for (const [idx, btn] of buttons.entries()) {
        btn.action = () => {
          set({ dialogProps: undefined });
          resolve(idx);
        };
      }
      onExit = () => {
        onExit?.();
        set({ dialogProps: undefined });
        reject();
      };
      set({ dialogProps: { buttons, title, text, onExit } });
    });
  },
}));
