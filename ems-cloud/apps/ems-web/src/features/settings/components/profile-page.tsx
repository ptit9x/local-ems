import {
  KeyRound,
  ShieldCheck,
} from "lucide-react";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { Input } from "../../../components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import { PageHeader } from "../../../components/dashboard/shared";

export function ProfilePage() {
  return (
    <>
      <PageHeader page="profile" actions={<Button>Save profile</Button>} />
      <div className="mt-7 grid gap-4 xl:grid-cols-[1.45fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Personal information</CardTitle>
            <CardDescription>
              Shown across your workspace and audit records
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="mb-6 flex flex-col gap-4 border-b border-border pb-6 sm:flex-row sm:items-center">
              <div className="grid size-20 place-items-center rounded-2xl bg-gradient-to-br from-emerald-700 to-emerald-500 font-display text-2xl font-semibold text-white shadow-lg">
                DA
              </div>
              <div className="flex-1">
                <div className="font-display text-lg font-semibold">
                  Demo Admin
                </div>
                <div className="mt-1 text-sm text-muted-foreground">
                  Administrator · CDI Energy Cloud
                </div>
                <Button variant="outline" size="sm" className="mt-3">
                  Change avatar
                </Button>
              </div>
            </div>
            <div className="grid gap-5 sm:grid-cols-2">
              <label className="space-y-2 text-sm font-medium">
                First name
                <Input defaultValue="Demo" />
              </label>
              <label className="space-y-2 text-sm font-medium">
                Last name
                <Input defaultValue="Admin" />
              </label>
              <label className="space-y-2 text-sm font-medium sm:col-span-2">
                Email address
                <Input type="email" defaultValue="admin@demo.local" />
              </label>
              <label className="space-y-2 text-sm font-medium">
                Job title
                <Input defaultValue="Energy Operations Manager" />
              </label>
              <label className="space-y-2 text-sm font-medium">
                Language
                <Select defaultValue="en">
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="vi">Tiếng Việt</SelectItem>
                    <SelectItem value="ja">日本語</SelectItem>
                  </SelectContent>
                </Select>
              </label>
            </div>
          </CardContent>
        </Card>
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Account security</CardTitle>
              <CardDescription>Keep your access protected</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center gap-3 rounded-xl border border-border p-4">
                <div className="grid size-10 place-items-center rounded-xl bg-emerald-50 text-emerald-700">
                  <KeyRound className="size-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium">Password</div>
                  <div className="text-xs text-muted-foreground">
                    Updated 24 days ago
                  </div>
                </div>
                <Button variant="outline" size="sm">
                  Change
                </Button>
              </div>
              <div className="flex items-center gap-3 rounded-xl border border-border p-4">
                <div className="grid size-10 place-items-center rounded-xl bg-violet-50 text-violet-700">
                  <ShieldCheck className="size-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium">
                    Two-factor authentication
                  </div>
                  <div className="text-xs text-emerald-700">Enabled</div>
                </div>
                <Button variant="outline" size="sm">
                  Manage
                </Button>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Active session</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-start gap-3">
                <div className="mt-1 size-2 rounded-full bg-emerald-500" />
                <div>
                  <div className="text-sm font-medium">Codex local preview</div>
                  <div className="mt-1 text-xs leading-5 text-muted-foreground">
                    Ho Chi Minh City · Active now
                    <br />
                    Chrome on macOS
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}
