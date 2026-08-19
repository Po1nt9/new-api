import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { DateTimePicker } from "@/components/datetime-picker";
import {
  SideDrawerSection,
  sideDrawerContentClassName,
  sideDrawerFooterClassName,
  sideDrawerFormClassName,
  sideDrawerHeaderClassName,
} from "@/components/drawer-layout";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { getCurrencyLabel } from "@/lib/currency";
import { getEditableQuotaStep } from "@/lib/format";
import { createInvitation, updateInvitation, getInvitation } from "../api";
import { SUCCESS_MESSAGES, ERROR_MESSAGES } from "../constants";
import {
  getInvitationFormSchema,
  type InvitationFormValues,
  INVITATION_FORM_DEFAULT_VALUES,
  transformFormDataToPayload,
  transformInvitationToFormDefaults,
} from "../lib";
import type { Invitation } from "../types";

type InvitationsMutateDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentRow?: Invitation;
};

export function InvitationsMutateDrawer(props: InvitationsMutateDrawerProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const isUpdate = !!props.currentRow;
  const invitationId = props.currentRow?.id;

  const form = useForm<InvitationFormValues>({
    resolver: zodResolver(getInvitationFormSchema(t)),
    defaultValues: INVITATION_FORM_DEFAULT_VALUES,
  });

  const countValue = useWatch({ control: form.control, name: "count" }) ?? 1;

  useEffect(() => {
    if (!props.open) return;
    if (!isUpdate || invitationId === undefined) {
      form.reset(INVITATION_FORM_DEFAULT_VALUES);
      return;
    }

    getInvitation(invitationId)
      .then((res) => {
        if (res.data) {
          form.reset(transformInvitationToFormDefaults(res.data));
        }
      })
      .catch(() => {
        if (props.currentRow) {
          form.reset(transformInvitationToFormDefaults(props.currentRow));
        }
      });
  }, [props.open, isUpdate, invitationId, props.currentRow, form]);

  const mutateMutation = useMutation({
    mutationFn: async (values: InvitationFormValues) => {
      const payload = transformFormDataToPayload(values);
      if (isUpdate && invitationId !== undefined) {
        return updateInvitation({ ...payload, id: invitationId });
      }
      return createInvitation(payload);
    },
    onSuccess: (res) => {
      if (res.success) {
        toast.success(
          t(isUpdate ? SUCCESS_MESSAGES.UPDATED : SUCCESS_MESSAGES.CREATED)
        );
        props.onOpenChange(false);
        queryClient.invalidateQueries({ queryKey: ["invitations"] });
      } else {
        toast.error(
          res.message ||
            t(
              isUpdate
                ? ERROR_MESSAGES.UPDATE_FAILED
                : ERROR_MESSAGES.CREATE_FAILED
            )
        );
      }
    },
    onError: () => {
      toast.error(
        t(
          isUpdate ? ERROR_MESSAGES.UPDATE_FAILED : ERROR_MESSAGES.CREATE_FAILED
        )
      );
    },
  });

  const onSubmit = (values: InvitationFormValues) => {
    mutateMutation.mutate(values);
  };

  let submitLabel = t("Generate");
  if (mutateMutation.isPending) {
    submitLabel = t("Saving...");
  } else if (isUpdate) {
    submitLabel = t("Update");
  }

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent className={sideDrawerContentClassName("sm:max-w-[600px]")}>
        <SheetHeader className={sideDrawerHeaderClassName()}>
          <SheetTitle>
            {isUpdate
              ? t("Edit Invitation Code")
              : t("Generate Invitation Codes")}
          </SheetTitle>
          <SheetDescription>
            {isUpdate
              ? t("Update the details of this invitation code.")
              : t(
                  "Generate one or multiple invitation codes for new user onboarding."
                )}
          </SheetDescription>
        </SheetHeader>

        <Form {...form}>
          <form
            id="invitation-form"
            onSubmit={form.handleSubmit(onSubmit)}
            className={sideDrawerFormClassName()}
          >
            <SideDrawerSection>
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("Name / Campaign")}</FormLabel>
                    <FormControl>
                      <Input placeholder={t("e.g. Beta Cohort 1")} {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {!isUpdate && (
                <>
                  <FormField
                    control={form.control}
                    name="count"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t("Generate Count")}</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            min={1}
                            max={100}
                            {...field}
                            value={field.value ?? 1}
                            onChange={(e) =>
                              field.onChange(Number(e.target.value))
                            }
                          />
                        </FormControl>
                        <FormDescription>
                          {t("Maximum 100 codes can be generated per batch.")}
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {countValue === 1 ? (
                    <FormField
                      control={form.control}
                      name="key"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>{t("Custom Code (Optional)")}</FormLabel>
                          <FormControl>
                            <Input
                              placeholder={t(
                                "e.g. VIP888 (leave empty to generate)"
                              )}
                              {...field}
                            />
                          </FormControl>
                          <FormDescription>
                            {t("Leave empty to generate a random code.")}
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  ) : (
                    <FormField
                      control={form.control}
                      name="prefix"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>{t("Custom Prefix (Optional)")}</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g. VIP-" {...field} />
                          </FormControl>
                          <FormDescription>
                            {t(
                              "Optional prefix prepended to generated random codes."
                            )}
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                </>
              )}

              <FormField
                control={form.control}
                name="quota_dollars"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t("Grant Quota ({{currency}})", {
                        currency: getCurrencyLabel(),
                      })}
                    </FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        step={getEditableQuotaStep()}
                        min={0}
                        {...field}
                        value={field.value ?? 0}
                        onChange={(e) => field.onChange(Number(e.target.value))}
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        "Initial quota balance granted to user upon registration."
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="group"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("Assign User Group (Optional)")}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. vip (leave empty for default)"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        "Target user group assigned upon successful registration."
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="expired_time"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("Expiration Date (Optional)")}</FormLabel>
                    <FormControl>
                      <DateTimePicker
                        value={field.value}
                        onChange={field.onChange}
                        placeholder={t(
                          "Leave empty for codes that never expire."
                        )}
                      />
                    </FormControl>
                    <FormDescription>
                      {t("Leave empty for codes that never expire.")}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </SideDrawerSection>
          </form>
        </Form>

        <SheetFooter className={sideDrawerFooterClassName()}>
          <SheetClose render={<Button variant="outline" />}>
            {t("Cancel")}
          </SheetClose>
          <Button
            form="invitation-form"
            type="submit"
            disabled={mutateMutation.isPending}
          >
            {submitLabel}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
