import { useTranslation } from "react-i18next";
import { SectionPageLayout } from "@/components/layout";
import { InvitationsDialogs } from "./components/invitations-dialogs";
import { InvitationsPrimaryButtons } from "./components/invitations-primary-buttons";
import { InvitationsProvider } from "./components/invitations-provider";
import { InvitationsTable } from "./components/invitations-table";

export function Invitations() {
  const { t } = useTranslation();
  return (
    <InvitationsProvider>
      <SectionPageLayout fixedContent>
        <SectionPageLayout.Title>
          {t("Invitation Codes")}
        </SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <InvitationsPrimaryButtons />
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <InvitationsTable />
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <InvitationsDialogs />
    </InvitationsProvider>
  );
}

export default Invitations;
