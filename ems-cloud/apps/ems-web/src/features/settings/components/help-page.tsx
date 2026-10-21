import {
  BookOpen,
  ChevronRight,
  Command,
  Mail,
  MessageCircle,
  Network,
  Search,
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
import { PageHeader } from "../../../components/dashboard/shared";

export function HelpPage() {
  const topics = [
    {
      title: "Getting started",
      detail: "Workspace setup, users and first site",
      icon: BookOpen,
    },
    {
      title: "Devices & telemetry",
      detail: "Provisioning, connectivity and data quality",
      icon: Network,
    },
    {
      title: "Commands & controls",
      detail: "Safe remote actions and audit trails",
      icon: Command,
    },
    {
      title: "Account & security",
      detail: "Roles, access and authentication",
      icon: ShieldCheck,
    },
  ];
  const questions = [
    "How do I add a new site?",
    "Why is a device showing as offline?",
    "How are commands acknowledged?",
    "How do I export an energy report?",
  ];
  return (
    <>
      <PageHeader page="help" />
      <div className="mt-7 rounded-2xl bg-[#0b1c18] p-5 text-white sm:p-8">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="font-display text-2xl font-semibold sm:text-3xl">
            Search the knowledge base
          </h2>
          <p className="mt-2 text-sm text-slate-400">
            Guides and answers for operating your energy fleet.
          </p>
          <div className="relative mt-6">
            <Search className="absolute left-4 top-1/2 size-5 -translate-y-1/2 text-slate-400" />
            <Input
              className="h-13 border-white/10 bg-white pl-12 text-slate-950 shadow-xl"
              placeholder="Search guides, devices, commands…"
            />
          </div>
        </div>
      </div>
      <div className="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {topics.map(({ title, detail, icon: Icon }) => (
          <Card
            key={title}
            className="group cursor-pointer transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-lg"
          >
            <CardContent className="p-5">
              <div className="flex items-start justify-between">
                <div className="grid size-11 place-items-center rounded-xl bg-emerald-50 text-emerald-700">
                  <Icon className="size-5" />
                </div>
                <ChevronRight className="size-4 text-slate-300 transition group-hover:translate-x-1 group-hover:text-emerald-600" />
              </div>
              <h3 className="mt-5 text-sm font-semibold">{title}</h3>
              <p className="mt-1 text-xs leading-5 text-muted-foreground">
                {detail}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-[1.45fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Frequently asked questions</CardTitle>
            <CardDescription>
              Quick answers from our support team
            </CardDescription>
          </CardHeader>
          <CardContent className="divide-y divide-border">
            {questions.map((question) => (
              <button
                key={question}
                className="flex w-full items-center justify-between gap-4 py-4 text-left text-sm font-medium hover:text-emerald-700"
              >
                {question}
                <ChevronRight className="size-4 shrink-0 text-slate-400" />
              </button>
            ))}
          </CardContent>
        </Card>
        <Card className="overflow-hidden">
          <CardContent className="flex h-full flex-col bg-gradient-to-br from-emerald-50 to-cyan-50 p-6">
            <div className="grid size-12 place-items-center rounded-2xl bg-white text-emerald-700 shadow-sm">
              <MessageCircle className="size-5" />
            </div>
            <h3 className="mt-6 font-display text-xl font-semibold">
              Still need help?
            </h3>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              Our energy operations support team is available around the clock.
            </p>
            <div className="mt-auto space-y-2 pt-8">
              <Button className="w-full">
                <MessageCircle className="size-4" />
                Start support chat
              </Button>
              <Button variant="outline" className="w-full bg-white">
                <Mail className="size-4" />
                Email support
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
